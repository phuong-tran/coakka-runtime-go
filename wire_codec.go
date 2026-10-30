package coakka_v2_connector

func cloneEnvelope(envelope *Envelope) *Envelope {
	if envelope == nil {
		return nil
	}
	clone := *envelope
	clone.Payload = append([]byte(nil), envelope.Payload...)
	if envelope.Headers != nil {
		clone.Headers = make(map[string]string, len(envelope.Headers))
		for key, value := range envelope.Headers {
			clone.Headers[key] = value
		}
	}
	return &clone
}
