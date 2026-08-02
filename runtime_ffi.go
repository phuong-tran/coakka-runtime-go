package coakka_v2_connector

/*
#cgo linux LDFLAGS: -ldl
#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <io.h>
#else
#include <dlfcn.h>
#include <poll.h>
#include <unistd.h>
#endif
#include <errno.h>
#include <limits.h>
#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct coakka_v2_runtime_t coakka_v2_runtime_t;

typedef struct {
  const char *system_name;
  const char *node_id;
  int strict_no_drop;
  int queue_capacity;
} coakka_v2_runtime_config_t;

typedef struct {
  size_t struct_size;
  uint32_t abi_version;
  uint32_t feature_flags;
  const char *runtime_version;
  const char *git_commit;
  const char *southbound_backend;
  const char *allocator_backend;
  const char *docs_hint;
} coakka_v2_runtime_info_t;

typedef struct {
  size_t struct_size;
  const char *system_name;
  const char *node_id;
  int strict_no_drop;
  int queue_capacity;
  size_t request_max_frame_size;
  size_t local_dispatch_batch_limit;
  int runtime_state;
  uint32_t snapshot_present;
  uint64_t applied_generation;
  size_t route_count;
  const char *southbound_bind_host;
  uint16_t southbound_bind_port;
  uint32_t configured_ingress_overload_mode;
  uint32_t configured_local_delivery_overload_mode;
  uint32_t configured_remote_outbound_overload_mode;
  size_t configured_remote_outbound_reply_reserve_slots;
  uint32_t effective_ingress_overload_mode;
  uint32_t effective_local_delivery_overload_mode;
  uint32_t effective_remote_outbound_overload_mode;
  size_t effective_remote_outbound_reply_reserve_slots;
} coakka_v2_runtime_config_view_t;

typedef struct {
  size_t struct_size;
  uint32_t flags;
  int request_write_fd;
  int response_read_fd;
  int deadletter_read_fd;
  int control_write_fd;
  int monitor_read_fd;
  int delivered_request_read_fd;
} coakka_v2_host_handles_t;

typedef struct {
  size_t struct_size;
  int runtime_state;
  uint32_t flags;
  uint64_t applied_generation;
} coakka_v2_runtime_health_t;

typedef struct {
  size_t struct_size;
  uint64_t applied_generation;
  size_t route_count;
  int runtime_state;
  size_t ingress_queue_capacity;
  size_t ingress_queue_depth;
  size_t ingress_queue_high_watermark;
  uint64_t queue_rejected_count;
  uint64_t route_miss_count;
  uint64_t deadletter_count;
  uint64_t delivery_failed_count;
  uint64_t caf_send_failed_count;
  uint64_t southbound_submit_attempt_count;
  uint64_t southbound_probe_connect_success_count;
  uint64_t southbound_probe_connect_failure_count;
  size_t request_max_frame_size;
  size_t local_dispatch_batch_limit;
  size_t delivered_request_outbound_queue_capacity;
  size_t delivered_request_outbound_queue_depth;
  size_t delivered_request_outbound_queue_high_watermark;
  uint64_t delivered_request_outbound_enqueue_block_count;
  size_t response_outbound_queue_capacity;
  size_t response_outbound_queue_depth;
  size_t response_outbound_queue_high_watermark;
  uint64_t response_outbound_enqueue_block_count;
  size_t deadletter_outbound_queue_capacity;
  size_t deadletter_outbound_queue_depth;
  size_t deadletter_outbound_queue_high_watermark;
  uint64_t deadletter_outbound_enqueue_block_count;
  uint64_t remote_reply_timeout_count;
  uint64_t late_remote_reply_drop_count;
  size_t remote_outbound_queue_capacity;
  size_t remote_outbound_queue_depth;
  size_t remote_outbound_queue_high_watermark;
  uint64_t remote_outbound_queue_rejected_count;
  uint64_t remote_outbound_expired_drop_count;
  uint64_t endpoint_unavailable_count;
  uint64_t remote_response_forwarded_count;
  uint64_t remote_deadletter_forwarded_count;
  size_t drained_route_count;
  uint64_t control_rejected_count;
  uint64_t control_invalid_count;
  uint64_t control_stale_generation_count;
  uint64_t control_bad_state_count;
  uint64_t control_io_count;
  size_t remote_outbound_reply_reserve_slots;
  uint64_t remote_outbound_reply_reservation_reject_count;
  uint32_t ingress_overload_mode;
  uint32_t local_delivery_overload_mode;
  uint32_t remote_outbound_overload_mode;
  uint64_t monitor_event_emitted_count;
  uint64_t monitor_event_dropped_count;
  uint64_t monitor_event_emitted_lifetime_count;
  uint64_t monitor_event_dropped_lifetime_count;
  size_t local_work_queue_capacity;
  size_t local_work_queue_depth;
  size_t local_work_queue_high_watermark;
  uint64_t delivered_request_outbound_direct_write_count;
  uint64_t response_outbound_direct_write_count;
  uint64_t deadletter_outbound_direct_write_count;
  uint64_t remote_outbound_one_way_drop_count;
} coakka_v2_runtime_stats_t;

typedef struct {
  size_t struct_size;
  uint32_t edition;
  uint32_t license_status;
  uint64_t compiled_capabilities;
  uint64_t entitled_capabilities;
  uint64_t effective_capabilities;
  uint32_t tcp_connection_defaults_revision;
  uint32_t reserved;
} coakka_v2_runtime_capability_snapshot_t;

typedef struct {
  size_t struct_size;
  uint64_t fields;
  uint32_t mode;
  uint32_t reserved;
  uint32_t max_connections;
  uint32_t reserved2;
  uint64_t max_requests_per_connection;
  uint64_t idle_timeout_ms;
} coakka_v2_tcp_connection_options_t;

typedef struct {
  size_t struct_size;
  uint32_t code;
  uint32_t reserved;
  uint64_t field;
  uint64_t minimum_value;
  uint64_t maximum_value;
} coakka_v2_tcp_connection_validation_t;

typedef struct {
  size_t struct_size;
  uint32_t defaults_revision;
  uint32_t mode;
  uint64_t applicable_fields;
  uint64_t explicitly_configured_fields;
  uint64_t defaulted_fields;
  uint64_t configurable_fields;
  uint32_t max_connections;
  uint32_t reserved;
  uint64_t max_requests_per_connection;
  uint64_t idle_timeout_ms;
} coakka_v2_tcp_connection_config_snapshot_t;

typedef struct {
  size_t struct_size;
  uint64_t fields;
  uint32_t mode;
  uint32_t credential_source_kind;
  uint32_t reload_mode;
  uint32_t reserved;
  uint64_t credential_generation;
  const char *credential_id;
  const char *ca_certificate_file;
  const char *identity_certificate_file;
  const char *private_key_file;
} coakka_v2_tcp_security_options_t;

typedef struct {
  size_t struct_size;
  uint32_t code;
  uint32_t reserved;
  uint64_t field;
} coakka_v2_tcp_security_validation_t;

typedef struct {
  size_t struct_size;
  uint32_t mode;
  uint32_t credential_source_kind;
  uint32_t reload_mode;
  uint32_t reload_status;
  uint64_t credential_generation;
  const char *credential_id;
} coakka_v2_tcp_security_config_snapshot_t;

typedef struct {
  size_t struct_size;
  uint32_t minimum_protocol_version;
  uint32_t maximum_protocol_version;
  uint64_t inbound_verification_flags;
  uint64_t outbound_verification_flags;
  int64_t identity_not_before_unix_seconds;
  int64_t identity_not_after_unix_seconds;
  char credential_id_value[128];
  char identity_fingerprint_sha256[65];
} coakka_v2_tcp_security_identity_info_t;

typedef struct {
  size_t struct_size;
  coakka_v2_tcp_security_config_snapshot_t config;
  coakka_v2_tcp_security_identity_info_t identity;
} coakka_v2_tcp_security_info_t;

typedef struct {
  size_t struct_size;
  int32_t apply_status;
  uint32_t changed;
  uint32_t reason;
  uint32_t runtime_state;
  coakka_v2_tcp_connection_validation_t validation;
  coakka_v2_tcp_connection_config_snapshot_t effective_config;
} coakka_v2_tcp_connection_apply_result_t;

typedef struct {
  size_t struct_size;
  int32_t apply_status;
  uint32_t changed;
  uint32_t reason;
  uint32_t runtime_state;
  coakka_v2_tcp_security_validation_t validation;
  coakka_v2_tcp_security_info_t active_security;
} coakka_v2_tcp_security_apply_result_t;

typedef uint32_t (*coakka_v2_runtime_get_abi_version_fn)(void);
typedef int (*coakka_v2_runtime_get_info_fn)(coakka_v2_runtime_info_t *out_info);
typedef int (*coakka_v2_runtime_get_config_fn)(coakka_v2_runtime_t *rt, coakka_v2_runtime_config_view_t *out_config);
typedef int (*coakka_v2_runtime_get_capabilities_fn)(coakka_v2_runtime_capability_snapshot_t *out_capabilities);
typedef coakka_v2_runtime_t *(*coakka_v2_runtime_create_fn)(const coakka_v2_runtime_config_t *cfg);
typedef void (*coakka_v2_runtime_destroy_fn)(coakka_v2_runtime_t *rt);
typedef int (*coakka_v2_runtime_get_host_handles_fn)(coakka_v2_runtime_t *rt, coakka_v2_host_handles_t *out_handles);
typedef int (*coakka_v2_runtime_start_fn)(coakka_v2_runtime_t *rt);
typedef int (*coakka_v2_runtime_stop_fn)(coakka_v2_runtime_t *rt);
typedef int (*coakka_v2_runtime_get_health_fn)(coakka_v2_runtime_t *rt, coakka_v2_runtime_health_t *out_health);
typedef int (*coakka_v2_runtime_get_stats_fn)(coakka_v2_runtime_t *rt, coakka_v2_runtime_stats_t *out_stats);
typedef int (*coakka_v2_runtime_submit_envelope_fn)(coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len);
typedef int (*coakka_v2_runtime_apply_control_envelope_fn)(coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len);
typedef int (*coakka_v2_monitor_consume_fn)(int fd, uint64_t *out_signal_count);
typedef int (*coakka_v2_runtime_apply_tcp_connection_options_ex_fn)(coakka_v2_runtime_t *rt, const coakka_v2_tcp_connection_options_t *options, coakka_v2_tcp_connection_apply_result_t *out_result);
typedef int (*coakka_v2_runtime_get_tcp_connection_config_fn)(coakka_v2_runtime_t *rt, coakka_v2_tcp_connection_config_snapshot_t *out_config);
typedef int (*coakka_v2_runtime_apply_tcp_security_options_ex_fn)(coakka_v2_runtime_t *rt, const coakka_v2_tcp_security_options_t *options, coakka_v2_tcp_security_apply_result_t *out_result);
typedef int (*coakka_v2_runtime_get_tcp_security_info_fn)(coakka_v2_runtime_t *rt, coakka_v2_tcp_security_info_t *out_info);
typedef const char *(*coakka_v2_transport_apply_reason_name_fn)(uint32_t reason);

typedef struct coakka_v2_go_bindings_t {
  void *handle;
  coakka_v2_runtime_get_abi_version_fn get_abi_version;
  coakka_v2_runtime_get_info_fn get_info;
  coakka_v2_runtime_get_config_fn get_config;
  coakka_v2_runtime_get_capabilities_fn get_capabilities;
  coakka_v2_runtime_create_fn create_runtime;
  coakka_v2_runtime_destroy_fn destroy_runtime;
  coakka_v2_runtime_get_host_handles_fn get_host_handles;
  coakka_v2_runtime_start_fn start_runtime;
  coakka_v2_runtime_stop_fn stop_runtime;
  coakka_v2_runtime_get_health_fn get_health;
  coakka_v2_runtime_get_stats_fn get_stats;
  coakka_v2_runtime_submit_envelope_fn submit_envelope;
  coakka_v2_runtime_apply_control_envelope_fn apply_control_envelope;
  coakka_v2_monitor_consume_fn monitor_consume;
  coakka_v2_runtime_apply_tcp_connection_options_ex_fn apply_tcp_connection_options_ex;
  coakka_v2_runtime_get_tcp_connection_config_fn get_tcp_connection_config;
  coakka_v2_runtime_apply_tcp_security_options_ex_fn apply_tcp_security_options_ex;
  coakka_v2_runtime_get_tcp_security_info_fn get_tcp_security_info;
  coakka_v2_transport_apply_reason_name_fn transport_apply_reason_name;
} coakka_v2_go_bindings_t;

static void *coakka_v2_go_process_library_handle = NULL;
static char *coakka_v2_go_process_library_path = NULL;

static char *coakka_v2_go_duplicate_text(const char *value) {
  if (value == NULL) {
    return NULL;
  }
  const size_t size = strlen(value) + 1u;
  char *copy = (char *)malloc(size);
  if (copy != NULL) {
    memcpy(copy, value, size);
  }
  return copy;
}

#ifdef _WIN32
static char *coakka_v2_go_windows_error(const char *operation) {
  char message[128];
  snprintf(message, sizeof(message), "%s failed win32_error=%lu", operation,
           (unsigned long)GetLastError());
  return coakka_v2_go_duplicate_text(message);
}
#endif

static int coakka_v2_go_load_symbol(void *handle, void **out_symbol, const char *symbol_name, char **error_out) {
#ifdef _WIN32
  FARPROC symbol = GetProcAddress((HMODULE)handle, symbol_name);
  if (symbol == NULL) {
    if (error_out != NULL) {
      *error_out = coakka_v2_go_windows_error("GetProcAddress");
    }
    return -1;
  }
  *out_symbol = (void *)(uintptr_t)symbol;
#else
  dlerror();
  *out_symbol = dlsym(handle, symbol_name);
  const char *error = dlerror();
  if (error != NULL) {
    if (error_out != NULL) {
      *error_out = coakka_v2_go_duplicate_text(error);
    }
    return -1;
  }
#endif
  return 0;
}

static coakka_v2_go_bindings_t *coakka_v2_go_open_library(const char *path, char **error_out) {
  int opened_new_library = 0;
  void *handle = NULL;
  if (coakka_v2_go_process_library_handle != NULL) {
    if (coakka_v2_go_process_library_path == NULL ||
        strcmp(coakka_v2_go_process_library_path, path) != 0) {
      if (error_out != NULL) {
        *error_out = coakka_v2_go_duplicate_text(
            "a different CoAkka runtime library is already loaded in this process");
      }
      return NULL;
    }
    handle = coakka_v2_go_process_library_handle;
  } else {
#ifdef _WIN32
    handle = (void *)LoadLibraryA(path);
#else
    handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
#endif
    if (handle == NULL) {
      if (error_out != NULL) {
#ifdef _WIN32
        *error_out = coakka_v2_go_windows_error("LoadLibraryA");
#else
        const char *error = dlerror();
        *error_out = error == NULL
            ? coakka_v2_go_duplicate_text("dlopen failed")
            : coakka_v2_go_duplicate_text(error);
#endif
      }
      return NULL;
    }
    coakka_v2_go_process_library_path = coakka_v2_go_duplicate_text(path);
    if (coakka_v2_go_process_library_path == NULL) {
      if (error_out != NULL) {
        *error_out = coakka_v2_go_duplicate_text("runtime path allocation failed");
      }
#ifdef _WIN32
      FreeLibrary((HMODULE)handle);
#else
      dlclose(handle);
#endif
      return NULL;
    }
    coakka_v2_go_process_library_handle = handle;
    opened_new_library = 1;
  }

  coakka_v2_go_bindings_t *bindings = (coakka_v2_go_bindings_t *)calloc(1, sizeof(coakka_v2_go_bindings_t));
  if (bindings == NULL) {
    if (error_out != NULL) {
      *error_out = coakka_v2_go_duplicate_text("calloc failed");
    }
    if (opened_new_library) {
#ifdef _WIN32
      FreeLibrary((HMODULE)handle);
#else
      dlclose(handle);
#endif
      free(coakka_v2_go_process_library_path);
      coakka_v2_go_process_library_path = NULL;
      coakka_v2_go_process_library_handle = NULL;
    }
    return NULL;
  }

  bindings->handle = handle;
  if (coakka_v2_go_load_symbol(handle, (void **)&bindings->get_abi_version, "coakka_v2_runtime_get_abi_version", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_info, "coakka_v2_runtime_get_info", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_config, "coakka_v2_runtime_get_config", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_capabilities, "coakka_v2_runtime_get_capabilities", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->create_runtime, "coakka_v2_runtime_create", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->destroy_runtime, "coakka_v2_runtime_destroy", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_host_handles, "coakka_v2_runtime_get_host_handles", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->start_runtime, "coakka_v2_runtime_start", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->stop_runtime, "coakka_v2_runtime_stop", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_health, "coakka_v2_runtime_get_health", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_stats, "coakka_v2_runtime_get_stats", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->submit_envelope, "coakka_v2_runtime_submit_envelope", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->apply_control_envelope, "coakka_v2_runtime_apply_control_envelope", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->monitor_consume, "coakka_v2_monitor_consume", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->apply_tcp_connection_options_ex, "coakka_v2_runtime_apply_tcp_connection_options_ex", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_tcp_connection_config, "coakka_v2_runtime_get_tcp_connection_config", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->apply_tcp_security_options_ex, "coakka_v2_runtime_apply_tcp_security_options_ex", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->get_tcp_security_info, "coakka_v2_runtime_get_tcp_security_info", error_out) != 0 ||
      coakka_v2_go_load_symbol(handle, (void **)&bindings->transport_apply_reason_name, "coakka_v2_transport_apply_reason_name", error_out) != 0) {
    if (opened_new_library) {
#ifdef _WIN32
      FreeLibrary((HMODULE)handle);
#else
      dlclose(handle);
#endif
      free(coakka_v2_go_process_library_path);
      coakka_v2_go_process_library_path = NULL;
      coakka_v2_go_process_library_handle = NULL;
    }
    free(bindings);
    return NULL;
  }

  return bindings;
}

static void coakka_v2_go_close_library(coakka_v2_go_bindings_t *bindings) {
  if (bindings == NULL) {
    return;
  }
  free(bindings);
}

static uint32_t coakka_v2_go_get_abi_version(coakka_v2_go_bindings_t *bindings) {
  return bindings->get_abi_version();
}

static int coakka_v2_go_get_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_info_t *out_info) {
  return bindings->get_info(out_info);
}

static int coakka_v2_go_get_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_config_view_t *out_config) {
  return bindings->get_config(rt, out_config);
}

static int coakka_v2_go_get_capabilities(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_capability_snapshot_t *out_capabilities) {
  return bindings->get_capabilities(out_capabilities);
}

static coakka_v2_runtime_t *coakka_v2_go_create_runtime(coakka_v2_go_bindings_t *bindings, const coakka_v2_runtime_config_t *cfg) {
  return bindings->create_runtime(cfg);
}

static void coakka_v2_go_destroy_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  bindings->destroy_runtime(rt);
}

static int coakka_v2_go_get_host_handles(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_host_handles_t *out_handles) {
  return bindings->get_host_handles(rt, out_handles);
}

static int coakka_v2_go_start_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  return bindings->start_runtime(rt);
}

static int coakka_v2_go_stop_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  return bindings->stop_runtime(rt);
}

static int coakka_v2_go_get_health(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_health_t *out_health) {
  return bindings->get_health(rt, out_health);
}

static int coakka_v2_go_get_stats(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_stats_t *out_stats) {
  return bindings->get_stats(rt, out_stats);
}

static int coakka_v2_go_submit_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len) {
  return bindings->submit_envelope(rt, buf, len);
}

static int coakka_v2_go_apply_control_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len) {
  return bindings->apply_control_envelope(rt, buf, len);
}

static int coakka_v2_go_monitor_consume(coakka_v2_go_bindings_t *bindings, int fd, uint64_t *out_signal_count) {
  return bindings->monitor_consume(fd, out_signal_count);
}

static int coakka_v2_go_apply_tcp_connection_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_connection_options_t *options, coakka_v2_tcp_connection_apply_result_t *out_result) {
  return bindings->apply_tcp_connection_options_ex(rt, options, out_result);
}

static int coakka_v2_go_get_tcp_connection_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_connection_config_snapshot_t *out_config) {
  return bindings->get_tcp_connection_config(rt, out_config);
}

static int coakka_v2_go_apply_tcp_security_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_security_options_t *options, coakka_v2_tcp_security_apply_result_t *out_result) {
  return bindings->apply_tcp_security_options_ex(rt, options, out_result);
}

static int coakka_v2_go_get_tcp_security_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_security_info_t *out_info) {
  return bindings->get_tcp_security_info(rt, out_info);
}

static const char *coakka_v2_go_transport_apply_reason_name(coakka_v2_go_bindings_t *bindings, uint32_t reason) {
  return bindings->transport_apply_reason_name(reason);
}

static int coakka_v2_go_wait_readable(int fd, int timeout_ms, int *out_ready) {
  if (out_ready == NULL) {
    return EINVAL;
  }
  *out_ready = 0;
#ifdef _WIN32
  const intptr_t raw_handle = _get_osfhandle(fd);
  if (raw_handle == -1) {
    return EBADF;
  }
  const DWORD wait_ms = timeout_ms < 0 ? INFINITE : (DWORD)timeout_ms;
  const DWORD result = WaitForSingleObject((HANDLE)raw_handle, wait_ms);
  if (result == WAIT_OBJECT_0) {
    *out_ready = 1;
    return 0;
  }
  if (result == WAIT_TIMEOUT) {
    return 0;
  }
  return EIO;
#else
  struct pollfd descriptor;
  memset(&descriptor, 0, sizeof(descriptor));
  descriptor.fd = fd;
  descriptor.events = POLLIN;
  const int result = poll(&descriptor, 1, timeout_ms);
  if (result < 0) {
    return errno;
  }
  if (result == 0) {
    return 0;
  }
  if ((descriptor.revents & (POLLERR | POLLNVAL)) != 0) {
    return EIO;
  }
  *out_ready = (descriptor.revents & (POLLIN | POLLHUP)) != 0;
  return 0;
#endif
}

static int coakka_v2_go_read_fd(int fd, uint8_t *buffer, size_t length,
                                size_t *out_read) {
  if (out_read == NULL || (buffer == NULL && length != 0u)) {
    return EINVAL;
  }
  *out_read = 0u;
#ifdef _WIN32
  const unsigned int bounded_length =
      length > (size_t)INT_MAX ? (unsigned int)INT_MAX : (unsigned int)length;
  const int result = _read(fd, buffer, bounded_length);
  if (result < 0) {
    return errno;
  }
  *out_read = (size_t)result;
#else
  const ssize_t result = read(fd, buffer, length);
  if (result < 0) {
    return errno;
  }
  *out_read = (size_t)result;
#endif
  return 0;
}

static int coakka_v2_go_close_fd(int fd) {
#ifdef _WIN32
  return _close(fd) == 0 ? 0 : errno;
#else
  return close(fd) == 0 ? 0 : errno;
#endif
}

static int coakka_v2_go_fd_error_retryable(int error_code) {
  return error_code == EINTR || error_code == EAGAIN
#ifdef EWOULDBLOCK
      || error_code == EWOULDBLOCK
#endif
      ;
}

static int coakka_v2_go_fd_error_terminal(int error_code) {
  return error_code == EBADF || error_code == EINVAL;
}
*/
import "C"

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const (
	COAKKAOK                                = 0
	COAKKAABIVersion                        = 1
	HostHandlesEnableMonitor         uint32 = 1 << 0
	HostHandlesSeparateDeliveredLane uint32 = 1 << 1
)

type RuntimeInfoSnapshot struct {
	AbiVersion        uint32 `json:"abiVersion"`
	FeatureFlags      uint32 `json:"featureFlags"`
	RuntimeVersion    string `json:"runtimeVersion"`
	GitCommit         string `json:"gitCommit"`
	SouthboundBackend string `json:"southboundBackend"`
	AllocatorBackend  string `json:"allocatorBackend"`
	DocsHint          string `json:"docsHint"`
}

type RuntimeConfigSnapshot struct {
	SystemName                               string `json:"systemName"`
	NodeID                                   string `json:"nodeId"`
	RuntimeState                             int32  `json:"runtimeState"`
	SnapshotPresent                          bool   `json:"snapshotPresent"`
	AppliedGeneration                        uint64 `json:"appliedGeneration"`
	RouteCount                               uint64 `json:"routeCount"`
	SouthboundBindHost                       string `json:"southboundBindHost"`
	SouthboundBindPort                       uint16 `json:"southboundBindPort"`
	EffectiveIngressOverloadMode             uint32 `json:"effectiveIngressOverloadMode"`
	EffectiveLocalDeliveryOverloadMode       uint32 `json:"effectiveLocalDeliveryOverloadMode"`
	EffectiveRemoteOutboundOverloadMode      uint32 `json:"effectiveRemoteOutboundOverloadMode"`
	EffectiveRemoteOutboundReplyReserveSlots uint64 `json:"effectiveRemoteOutboundReplyReserveSlots"`
}

type RuntimeHealthSnapshot struct {
	RuntimeState      int32  `json:"runtimeState"`
	Flags             uint32 `json:"flags"`
	AppliedGeneration uint64 `json:"appliedGeneration"`
}

type RuntimeStatsSnapshot struct {
	AppliedGeneration                         uint64 `json:"appliedGeneration"`
	RouteCount                                uint64 `json:"routeCount"`
	RuntimeState                              int32  `json:"runtimeState"`
	QueueRejectedCount                        uint64 `json:"queueRejectedCount"`
	RouteMissCount                            uint64 `json:"routeMissCount"`
	DeadletterCount                           uint64 `json:"deadletterCount"`
	DeliveryFailedCount                       uint64 `json:"deliveryFailedCount"`
	ControlRejectedCount                      uint64 `json:"controlRejectedCount"`
	MonitorEventEmittedCount                  uint64 `json:"monitorEventEmittedCount"`
	MonitorEventDroppedCount                  uint64 `json:"monitorEventDroppedCount"`
	MonitorEventEmittedLifetimeCount          uint64 `json:"monitorEventEmittedLifetimeCount"`
	MonitorEventDroppedLifetimeCount          uint64 `json:"monitorEventDroppedLifetimeCount"`
	LocalWorkQueueCapacity                    uint64 `json:"localWorkQueueCapacity"`
	LocalWorkQueueDepth                       uint64 `json:"localWorkQueueDepth"`
	LocalWorkQueueHighWatermark               uint64 `json:"localWorkQueueHighWatermark"`
	DeliveredRequestOutboundDirectWriteCount  uint64 `json:"deliveredRequestOutboundDirectWriteCount"`
	ResponseOutboundDirectWriteCount          uint64 `json:"responseOutboundDirectWriteCount"`
	DeadletterOutboundDirectWriteCount        uint64 `json:"deadletterOutboundDirectWriteCount"`
	RemoteOutboundQueueCapacity               uint64 `json:"remoteOutboundQueueCapacity"`
	RemoteOutboundQueueDepth                  uint64 `json:"remoteOutboundQueueDepth"`
	RemoteOutboundQueueHighWatermark          uint64 `json:"remoteOutboundQueueHighWatermark"`
	RemoteOutboundQueueRejectedCount          uint64 `json:"remoteOutboundQueueRejectedCount"`
	RemoteOutboundExpiredDropCount            uint64 `json:"remoteOutboundExpiredDropCount"`
	RemoteOutboundReplyReserveSlots           uint64 `json:"remoteOutboundReplyReserveSlots"`
	RemoteOutboundReplyReservationRejectCount uint64 `json:"remoteOutboundReplyReservationRejectCount"`
	RemoteOutboundOneWayDropCount             uint64 `json:"remoteOutboundOneWayDropCount"`
}

type HostHandlesSnapshot struct {
	Flags                  uint32
	RequestWriteFD         int
	ResponseReadFD         int
	DeadletterReadFD       int
	ControlWriteFD         int
	MonitorReadFD          int
	DeliveredRequestReadFD int
}

type nativeBindings struct {
	ptr *C.coakka_v2_go_bindings_t
}

type nativeRuntime unsafe.Pointer

var nativeLibraryMu sync.Mutex

func openNativeBindings(path string) (*nativeBindings, error) {
	nativeLibraryMu.Lock()
	defer nativeLibraryMu.Unlock()
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cError *C.char
	ptr := C.coakka_v2_go_open_library(cPath, &cError)
	if cError != nil {
		defer C.free(unsafe.Pointer(cError))
	}
	if ptr == nil {
		if cError != nil {
			return nil, fmt.Errorf("load runtime library: %s", C.GoString(cError))
		}
		return nil, fmt.Errorf("load runtime library: unknown error")
	}
	return &nativeBindings{ptr: ptr}, nil
}

func (b *nativeBindings) close() {
	if b == nil || b.ptr == nil {
		return
	}
	C.coakka_v2_go_close_library(b.ptr)
	b.ptr = nil
}

func (b *nativeBindings) getAbiVersion() uint32 {
	return uint32(C.coakka_v2_go_get_abi_version(b.ptr))
}

func (b *nativeBindings) readRuntimeInfo() (RuntimeInfoSnapshot, error) {
	info := C.coakka_v2_runtime_info_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_info_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_info(b.ptr, &info), "runtime_get_info"); err != nil {
		return RuntimeInfoSnapshot{}, err
	}
	return RuntimeInfoSnapshot{
		AbiVersion:        uint32(info.abi_version),
		FeatureFlags:      uint32(info.feature_flags),
		RuntimeVersion:    cString(info.runtime_version),
		GitCommit:         cString(info.git_commit),
		SouthboundBackend: cString(info.southbound_backend),
		AllocatorBackend:  cString(info.allocator_backend),
		DocsHint:          cString(info.docs_hint),
	}, nil
}

func (b *nativeBindings) readRuntimeCapabilities() (RuntimeCapabilitiesSnapshot, error) {
	native := C.coakka_v2_runtime_capability_snapshot_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_capability_snapshot_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_capabilities(b.ptr, &native),
		"runtime_get_capabilities",
	); err != nil {
		return RuntimeCapabilitiesSnapshot{}, err
	}
	return RuntimeCapabilitiesSnapshot{
		Edition:                       uint32(native.edition),
		LicenseStatus:                 uint32(native.license_status),
		CompiledCapabilities:          RuntimeCapability(native.compiled_capabilities),
		EntitledCapabilities:          RuntimeCapability(native.entitled_capabilities),
		EffectiveCapabilities:         RuntimeCapability(native.effective_capabilities),
		TCPConnectionDefaultsRevision: uint32(native.tcp_connection_defaults_revision),
	}, nil
}

func (b *nativeBindings) createRuntime(config ConnectorConfig) (nativeRuntime, error) {
	systemName := C.CString(config.SystemName)
	nodeID := C.CString(config.NodeID)
	defer C.free(unsafe.Pointer(systemName))
	defer C.free(unsafe.Pointer(nodeID))
	cfg := C.coakka_v2_runtime_config_t{
		system_name:    systemName,
		node_id:        nodeID,
		strict_no_drop: boolToCInt(config.StrictNoDrop),
		queue_capacity: C.int(config.QueueCapacity),
	}
	runtime := C.coakka_v2_go_create_runtime(b.ptr, &cfg)
	if runtime == nil {
		return nil, fmt.Errorf("coakka_v2_runtime_create returned null")
	}
	return nativeRuntime(runtime), nil
}

func (b *nativeBindings) destroyRuntime(runtime nativeRuntime) {
	if runtime != nil {
		C.coakka_v2_go_destroy_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime))
	}
}

func (b *nativeBindings) getHostHandles(runtime nativeRuntime, flags uint32) (HostHandlesSnapshot, error) {
	handles := C.coakka_v2_host_handles_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_host_handles_t),
		flags:       C.uint32_t(flags),
	}
	if err := requireStatus(C.coakka_v2_go_get_host_handles(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &handles), "get_host_handles"); err != nil {
		return HostHandlesSnapshot{}, err
	}
	return HostHandlesSnapshot{
		Flags:                  uint32(handles.flags),
		RequestWriteFD:         int(handles.request_write_fd),
		ResponseReadFD:         int(handles.response_read_fd),
		DeadletterReadFD:       int(handles.deadletter_read_fd),
		ControlWriteFD:         int(handles.control_write_fd),
		MonitorReadFD:          int(handles.monitor_read_fd),
		DeliveredRequestReadFD: int(handles.delivered_request_read_fd),
	}, nil
}

func (b *nativeBindings) startRuntime(runtime nativeRuntime) error {
	return requireStatus(C.coakka_v2_go_start_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime)), "runtime_start")
}

func (b *nativeBindings) stopRuntime(runtime nativeRuntime) error {
	return requireStatus(C.coakka_v2_go_stop_runtime(b.ptr, (*C.coakka_v2_runtime_t)(runtime)), "runtime_stop")
}

func (b *nativeBindings) getHealth(runtime nativeRuntime) (RuntimeHealthSnapshot, error) {
	health := C.coakka_v2_runtime_health_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_health_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_health(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &health), "runtime_get_health"); err != nil {
		return RuntimeHealthSnapshot{}, err
	}
	return RuntimeHealthSnapshot{
		RuntimeState:      int32(health.runtime_state),
		Flags:             uint32(health.flags),
		AppliedGeneration: uint64(health.applied_generation),
	}, nil
}

func (b *nativeBindings) getStats(runtime nativeRuntime) (RuntimeStatsSnapshot, error) {
	stats := C.coakka_v2_runtime_stats_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_stats_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_stats(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &stats), "runtime_get_stats"); err != nil {
		return RuntimeStatsSnapshot{}, err
	}
	return RuntimeStatsSnapshot{
		AppliedGeneration:                         uint64(stats.applied_generation),
		RouteCount:                                uint64(stats.route_count),
		RuntimeState:                              int32(stats.runtime_state),
		QueueRejectedCount:                        uint64(stats.queue_rejected_count),
		RouteMissCount:                            uint64(stats.route_miss_count),
		DeadletterCount:                           uint64(stats.deadletter_count),
		DeliveryFailedCount:                       uint64(stats.delivery_failed_count),
		ControlRejectedCount:                      uint64(stats.control_rejected_count),
		MonitorEventEmittedCount:                  uint64(stats.monitor_event_emitted_count),
		MonitorEventDroppedCount:                  uint64(stats.monitor_event_dropped_count),
		MonitorEventEmittedLifetimeCount:          uint64(stats.monitor_event_emitted_lifetime_count),
		MonitorEventDroppedLifetimeCount:          uint64(stats.monitor_event_dropped_lifetime_count),
		LocalWorkQueueCapacity:                    uint64(stats.local_work_queue_capacity),
		LocalWorkQueueDepth:                       uint64(stats.local_work_queue_depth),
		LocalWorkQueueHighWatermark:               uint64(stats.local_work_queue_high_watermark),
		DeliveredRequestOutboundDirectWriteCount:  uint64(stats.delivered_request_outbound_direct_write_count),
		ResponseOutboundDirectWriteCount:          uint64(stats.response_outbound_direct_write_count),
		DeadletterOutboundDirectWriteCount:        uint64(stats.deadletter_outbound_direct_write_count),
		RemoteOutboundQueueCapacity:               uint64(stats.remote_outbound_queue_capacity),
		RemoteOutboundQueueDepth:                  uint64(stats.remote_outbound_queue_depth),
		RemoteOutboundQueueHighWatermark:          uint64(stats.remote_outbound_queue_high_watermark),
		RemoteOutboundQueueRejectedCount:          uint64(stats.remote_outbound_queue_rejected_count),
		RemoteOutboundExpiredDropCount:            uint64(stats.remote_outbound_expired_drop_count),
		RemoteOutboundReplyReserveSlots:           uint64(stats.remote_outbound_reply_reserve_slots),
		RemoteOutboundReplyReservationRejectCount: uint64(stats.remote_outbound_reply_reservation_reject_count),
		RemoteOutboundOneWayDropCount:             uint64(stats.remote_outbound_one_way_drop_count),
	}, nil
}

func (b *nativeBindings) getConfig(runtime nativeRuntime) (RuntimeConfigSnapshot, error) {
	config := C.coakka_v2_runtime_config_view_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_runtime_config_view_t),
	}
	if err := requireStatus(C.coakka_v2_go_get_config(b.ptr, (*C.coakka_v2_runtime_t)(runtime), &config), "runtime_get_config"); err != nil {
		return RuntimeConfigSnapshot{}, err
	}
	return RuntimeConfigSnapshot{
		SystemName:                               cString(config.system_name),
		NodeID:                                   cString(config.node_id),
		RuntimeState:                             int32(config.runtime_state),
		SnapshotPresent:                          config.snapshot_present != 0,
		AppliedGeneration:                        uint64(config.applied_generation),
		RouteCount:                               uint64(config.route_count),
		SouthboundBindHost:                       cString(config.southbound_bind_host),
		SouthboundBindPort:                       uint16(config.southbound_bind_port),
		EffectiveIngressOverloadMode:             uint32(config.effective_ingress_overload_mode),
		EffectiveLocalDeliveryOverloadMode:       uint32(config.effective_local_delivery_overload_mode),
		EffectiveRemoteOutboundOverloadMode:      uint32(config.effective_remote_outbound_overload_mode),
		EffectiveRemoteOutboundReplyReserveSlots: uint64(config.effective_remote_outbound_reply_reserve_slots),
	}, nil
}

func (b *nativeBindings) applyTCPConnectionStrategy(
	runtime nativeRuntime,
	spec RuntimeTCPConnectionStrategySpec,
) (RuntimeTCPConnectionApplyResult, error) {
	options := C.coakka_v2_tcp_connection_options_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_options_t),
		fields:      C.uint64_t(transportConnectionFieldMode),
		mode:        C.uint32_t(spec.Mode),
	}
	if spec.MaxConnections != nil {
		options.fields |= C.uint64_t(transportConnectionFieldMaxConnection)
		options.max_connections = C.uint32_t(*spec.MaxConnections)
	}
	if spec.MaxRequestsPerConnection != nil {
		options.fields |= C.uint64_t(transportConnectionFieldMaxRequests)
		options.max_requests_per_connection = C.uint64_t(*spec.MaxRequestsPerConnection)
	}
	if spec.IdleTimeoutMillis != nil {
		options.fields |= C.uint64_t(transportConnectionFieldIdleTimeout)
		options.idle_timeout_ms = C.uint64_t(*spec.IdleTimeoutMillis)
	}
	native := C.coakka_v2_tcp_connection_apply_result_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_apply_result_t),
	}
	status := C.coakka_v2_go_apply_tcp_connection_options_ex(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&options,
		&native,
	)
	if int32(status) != int32(native.apply_status) {
		return RuntimeTCPConnectionApplyResult{}, fmt.Errorf(
			"tcp connection apply returned status=%d but result status=%d",
			int32(status),
			int32(native.apply_status),
		)
	}
	return RuntimeTCPConnectionApplyResult{
		Status:          CoakkaStatus(native.apply_status),
		Changed:         native.changed != 0,
		Reason:          RuntimeTransportApplyReason(native.reason),
		ReasonName:      cString(C.coakka_v2_go_transport_apply_reason_name(b.ptr, native.reason)),
		RuntimeState:    uint32(native.runtime_state),
		ValidationCode:  RuntimeTCPConnectionValidationCode(native.validation.code),
		ValidationField: uint64(native.validation.field),
		MinimumValue:    uint64(native.validation.minimum_value),
		MaximumValue:    uint64(native.validation.maximum_value),
		ActiveConfig:    connectionConfigSnapshot(native.effective_config),
	}, nil
}

func (b *nativeBindings) getTCPConnectionConfig(
	runtime nativeRuntime,
) (RuntimeTCPConnectionConfigSnapshot, error) {
	native := C.coakka_v2_tcp_connection_config_snapshot_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_connection_config_snapshot_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_tcp_connection_config(
			b.ptr,
			(*C.coakka_v2_runtime_t)(runtime),
			&native,
		),
		"runtime_get_tcp_connection_config",
	); err != nil {
		return RuntimeTCPConnectionConfigSnapshot{}, err
	}
	return connectionConfigSnapshot(native), nil
}

func (b *nativeBindings) applyTCPSecurity(
	runtime nativeRuntime,
	spec RuntimeTCPSecuritySpec,
) (RuntimeTCPSecurityApplyResult, error) {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "credentialId", value: spec.CredentialID},
		{name: "caCertificateFile", value: spec.CACertificateFile},
		{name: "identityCertificateFile", value: spec.IdentityCertificateFile},
		{name: "privateKeyFile", value: spec.PrivateKeyFile},
	} {
		if err := validateTransportText(field.value, field.name); err != nil {
			return RuntimeTCPSecurityApplyResult{}, err
		}
	}
	credentialID := C.CString(spec.CredentialID)
	caCertificateFile := C.CString(spec.CACertificateFile)
	identityCertificateFile := C.CString(spec.IdentityCertificateFile)
	privateKeyFile := C.CString(spec.PrivateKeyFile)
	defer C.free(unsafe.Pointer(credentialID))
	defer C.free(unsafe.Pointer(caCertificateFile))
	defer C.free(unsafe.Pointer(identityCertificateFile))
	defer C.free(unsafe.Pointer(privateKeyFile))

	options := C.coakka_v2_tcp_security_options_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_options_t),
		fields:      C.uint64_t(transportSecurityFieldMode),
		mode:        C.uint32_t(spec.Mode),
	}
	if spec.Mode != RuntimeTCPSecurityPlaintext {
		options.fields = C.uint64_t(transportSecurityAllFields)
		options.credential_source_kind = C.uint32_t(transportCredentialSourceFile)
		options.reload_mode = C.uint32_t(spec.ReloadMode)
		options.credential_generation = C.uint64_t(spec.CredentialGeneration)
		options.credential_id = credentialID
		options.ca_certificate_file = caCertificateFile
		options.identity_certificate_file = identityCertificateFile
		options.private_key_file = privateKeyFile
	} else {
		if spec.ReloadMode != RuntimeTLSReloadGraceful {
			options.fields |= C.uint64_t(transportSecurityFieldReloadMode)
			options.reload_mode = C.uint32_t(spec.ReloadMode)
		}
		if spec.CredentialGeneration != 0 {
			options.fields |= C.uint64_t(transportSecurityFieldGeneration)
			options.credential_generation = C.uint64_t(spec.CredentialGeneration)
		}
		if spec.CredentialID != "" {
			options.fields |= C.uint64_t(transportSecurityFieldCredentialID)
			options.credential_id = credentialID
		}
		if spec.CACertificateFile != "" || spec.IdentityCertificateFile != "" || spec.PrivateKeyFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldSource)
			options.credential_source_kind = C.uint32_t(transportCredentialSourceFile)
		}
		if spec.CACertificateFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldCAFile)
			options.ca_certificate_file = caCertificateFile
		}
		if spec.IdentityCertificateFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldIdentityFile)
			options.identity_certificate_file = identityCertificateFile
		}
		if spec.PrivateKeyFile != "" {
			options.fields |= C.uint64_t(transportSecurityFieldKeyFile)
			options.private_key_file = privateKeyFile
		}
	}
	native := C.coakka_v2_tcp_security_apply_result_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_apply_result_t),
	}
	status := C.coakka_v2_go_apply_tcp_security_options_ex(
		b.ptr,
		(*C.coakka_v2_runtime_t)(runtime),
		&options,
		&native,
	)
	if int32(status) != int32(native.apply_status) {
		return RuntimeTCPSecurityApplyResult{}, fmt.Errorf(
			"tcp security apply returned status=%d but result status=%d",
			int32(status),
			int32(native.apply_status),
		)
	}
	return RuntimeTCPSecurityApplyResult{
		Status:          CoakkaStatus(native.apply_status),
		Changed:         native.changed != 0,
		Reason:          RuntimeTransportApplyReason(native.reason),
		ReasonName:      cString(C.coakka_v2_go_transport_apply_reason_name(b.ptr, native.reason)),
		RuntimeState:    uint32(native.runtime_state),
		ValidationCode:  RuntimeTCPSecurityValidationCode(native.validation.code),
		ValidationField: uint64(native.validation.field),
		ActiveSecurity:  securityInfoSnapshot(native.active_security),
	}, nil
}

func (b *nativeBindings) getTCPSecurityInfo(
	runtime nativeRuntime,
) (RuntimeTCPSecurityInfoSnapshot, error) {
	native := C.coakka_v2_tcp_security_info_t{
		struct_size: C.size_t(C.sizeof_coakka_v2_tcp_security_info_t),
	}
	if err := requireStatus(
		C.coakka_v2_go_get_tcp_security_info(
			b.ptr,
			(*C.coakka_v2_runtime_t)(runtime),
			&native,
		),
		"runtime_get_tcp_security_info",
	); err != nil {
		return RuntimeTCPSecurityInfoSnapshot{}, err
	}
	return securityInfoSnapshot(native), nil
}

func connectionConfigSnapshot(
	native C.coakka_v2_tcp_connection_config_snapshot_t,
) RuntimeTCPConnectionConfigSnapshot {
	return RuntimeTCPConnectionConfigSnapshot{
		DefaultsRevision:           uint32(native.defaults_revision),
		Mode:                       RuntimeTCPConnectionMode(native.mode),
		ApplicableFields:           uint64(native.applicable_fields),
		ExplicitlyConfiguredFields: uint64(native.explicitly_configured_fields),
		DefaultedFields:            uint64(native.defaulted_fields),
		ConfigurableFields:         uint64(native.configurable_fields),
		MaxConnections:             uint32(native.max_connections),
		MaxRequestsPerConnection:   uint64(native.max_requests_per_connection),
		IdleTimeoutMillis:          uint64(native.idle_timeout_ms),
	}
}

func securityInfoSnapshot(native C.coakka_v2_tcp_security_info_t) RuntimeTCPSecurityInfoSnapshot {
	return RuntimeTCPSecurityInfoSnapshot{
		Mode:                         RuntimeTCPSecurityMode(native.config.mode),
		CredentialSourceKind:         uint32(native.config.credential_source_kind),
		ReloadMode:                   RuntimeTLSReloadMode(native.config.reload_mode),
		ReloadStatus:                 uint32(native.config.reload_status),
		CredentialGeneration:         uint64(native.config.credential_generation),
		CredentialID:                 C.GoString(&native.identity.credential_id_value[0]),
		MinimumProtocolVersion:       uint32(native.identity.minimum_protocol_version),
		MaximumProtocolVersion:       uint32(native.identity.maximum_protocol_version),
		InboundVerificationFlags:     uint64(native.identity.inbound_verification_flags),
		OutboundVerificationFlags:    uint64(native.identity.outbound_verification_flags),
		IdentityNotBeforeUnixSeconds: int64(native.identity.identity_not_before_unix_seconds),
		IdentityNotAfterUnixSeconds:  int64(native.identity.identity_not_after_unix_seconds),
		IdentityFingerprintSHA256:    C.GoString(&native.identity.identity_fingerprint_sha256[0]),
	}
}

func transportABISizes() RuntimeTransportABISizes {
	return RuntimeTransportABISizes{
		Capabilities:         uintptr(C.sizeof_coakka_v2_runtime_capability_snapshot_t),
		ConnectionOptions:    uintptr(C.sizeof_coakka_v2_tcp_connection_options_t),
		ConnectionValidation: uintptr(C.sizeof_coakka_v2_tcp_connection_validation_t),
		ConnectionConfig:     uintptr(C.sizeof_coakka_v2_tcp_connection_config_snapshot_t),
		ConnectionApply:      uintptr(C.sizeof_coakka_v2_tcp_connection_apply_result_t),
		SecurityOptions:      uintptr(C.sizeof_coakka_v2_tcp_security_options_t),
		SecurityValidation:   uintptr(C.sizeof_coakka_v2_tcp_security_validation_t),
		SecurityConfig:       uintptr(C.sizeof_coakka_v2_tcp_security_config_snapshot_t),
		SecurityIdentity:     uintptr(C.sizeof_coakka_v2_tcp_security_identity_info_t),
		SecurityInfo:         uintptr(C.sizeof_coakka_v2_tcp_security_info_t),
		SecurityApply:        uintptr(C.sizeof_coakka_v2_tcp_security_apply_result_t),
	}
}

func (b *nativeBindings) submitEnvelope(runtime nativeRuntime, bytes []byte) error {
	return requireStatus(C.coakka_v2_go_submit_envelope(b.ptr, (*C.coakka_v2_runtime_t)(runtime), bytePointer(bytes), C.size_t(len(bytes))), "runtime_submit_envelope")
}

func (b *nativeBindings) applyControlEnvelope(runtime nativeRuntime, bytes []byte) error {
	return requireStatus(C.coakka_v2_go_apply_control_envelope(b.ptr, (*C.coakka_v2_runtime_t)(runtime), bytePointer(bytes), C.size_t(len(bytes))), "runtime_apply_control_envelope")
}

func (b *nativeBindings) monitorConsume(fd int) (uint64, error) {
	var signalCount C.uint64_t
	if err := requireStatus(C.coakka_v2_go_monitor_consume(b.ptr, C.int(fd), &signalCount), "monitor_consume"); err != nil {
		return 0, err
	}
	return uint64(signalCount), nil
}

func requireStatus(status C.int, operation string) error {
	if int(status) == COAKKAOK {
		return nil
	}
	return fmt.Errorf("%s failed status=%d", operation, int(status))
}

func cString(ptr *C.char) string {
	if ptr == nil {
		return ""
	}
	return C.GoString(ptr)
}

func boolToCInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}

func bytePointer(bytes []byte) *C.uint8_t {
	if len(bytes) == 0 {
		return nil
	}
	return (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(bytes)))
}

type nativeFDError struct {
	operation string
	code      int
}

func (e *nativeFDError) Error() string {
	return fmt.Sprintf("%s failed errno=%d", e.operation, e.code)
}

func waitReadable(fd int, timeout time.Duration) (bool, error) {
	timeoutMillis := int64(timeout / time.Millisecond)
	if timeout < 0 {
		timeoutMillis = -1
	} else if timeoutMillis > 1<<31-1 {
		timeoutMillis = 1<<31 - 1
	}
	var ready C.int
	code := C.coakka_v2_go_wait_readable(C.int(fd), C.int(timeoutMillis), &ready)
	if code != 0 {
		return false, &nativeFDError{operation: "wait readable", code: int(code)}
	}
	return ready != 0, nil
}

func nativeReadFD(fd int, buffer []byte) (int, error) {
	var readCount C.size_t
	code := C.coakka_v2_go_read_fd(
		C.int(fd),
		bytePointer(buffer),
		C.size_t(len(buffer)),
		&readCount,
	)
	if code != 0 {
		return 0, &nativeFDError{operation: "read", code: int(code)}
	}
	return int(readCount), nil
}

func isRetryableReadError(err error) bool {
	fdError, ok := err.(*nativeFDError)
	return ok && C.coakka_v2_go_fd_error_retryable(C.int(fdError.code)) != 0
}

func isTerminalReadError(err error) bool {
	fdError, ok := err.(*nativeFDError)
	return ok && C.coakka_v2_go_fd_error_terminal(C.int(fdError.code)) != 0
}

func safeCloseFD(fd int) {
	if fd >= 0 {
		C.coakka_v2_go_close_fd(C.int(fd))
	}
}
