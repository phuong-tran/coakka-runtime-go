package coakka_v2_connector

import "sync"

// runtimeLifetime owns Go admission and pointer retirement, not native state.
// It is embedded by value and is never copied after the first acquired call.
type runtimeLifetime struct {
	admission sync.Mutex
	closeMu   sync.Mutex
	calls     sync.WaitGroup
	closed    chan struct{}
	closing   bool
	retired   bool
}

func (l *runtimeLifetime) acquire() bool {
	l.admission.Lock()
	defer l.admission.Unlock()
	if l.closing {
		return false
	}
	l.calls.Add(1)
	return true
}

func (l *runtimeLifetime) release() { l.calls.Done() }

// stop must release its native-operation lock before returning: admitted calls
// may still need that lock before they can release their borrow. A refusal
// leaves admission closed but retains every owner for a later Close retry.
func (l *runtimeLifetime) close(stop func() error, joinHost func(), retire func()) error {
	l.closeMu.Lock()
	defer l.closeMu.Unlock()
	if l.retired {
		return nil
	}
	l.admission.Lock()
	if !l.closing {
		l.closing = true
		close(l.closed)
	}
	l.admission.Unlock()
	if err := stop(); err != nil {
		return err
	}
	l.calls.Wait()
	// Accepted registration can finish creating its Go workers before Wait
	// begins. No Add-after-Wait race and no native-operation lock is retained.
	joinHost()
	retire()
	l.retired = true
	return nil
}
