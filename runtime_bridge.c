#include "runtime_bridge.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#else
#include <dlfcn.h>
#endif

/* Apply strict owned-C diagnostics here, not to cgo-generated probe bodies. */
#if defined(__GNUC__)
#pragma GCC diagnostic error "-Wextra"
#pragma GCC diagnostic error "-Wpedantic"
#endif

/* Function types follow the sealed Core declarations. Every use is checked
 * against that declaration; these are not independently maintained ABI types. */
typedef uint32_t (*coakka_v2_runtime_get_abi_version_fn)(void);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_info_fn)(coakka_v2_runtime_info_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_config_fn)(coakka_v2_runtime_t *, coakka_v2_runtime_config_view_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_capabilities_fn)(coakka_v2_runtime_capability_snapshot_t *);
typedef coakka_v2_runtime_t *(*coakka_v2_runtime_create_fn)(const coakka_v2_runtime_config_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_create_native_fn)(const coakka_v2_runtime_config_t *, const coakka_v2_runtime_startup_options_t *, coakka_v2_runtime_t **);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_startup_config_fn)(coakka_v2_runtime_t *, coakka_v2_runtime_startup_config_t *);
typedef void (*coakka_v2_runtime_destroy_fn)(coakka_v2_runtime_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_apply_network_options_fn)(coakka_v2_runtime_t *, const coakka_v2_network_options_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_network_config_fn)(coakka_v2_runtime_t *, coakka_v2_network_config_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_client_encode_message_value_fn)(const coakka_v2_message_value_t *, uint8_t **, size_t *);
typedef coakka_v2_status_t (*coakka_v2_client_decode_message_value_fn)(const uint8_t *, size_t, coakka_v2_message_value_owner_t **, coakka_v2_message_value_t *);
typedef coakka_v2_status_t (*coakka_v2_client_decode_deadletter_value_fn)(const uint8_t *, size_t, coakka_v2_message_value_owner_t **, coakka_v2_deadletter_value_t *);
typedef void (*coakka_v2_message_value_release_fn)(coakka_v2_message_value_owner_t *);
typedef coakka_v2_status_t (*coakka_v2_client_encode_control_value_fn)(const coakka_v2_control_value_t *, uint8_t **, size_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_host_handles_fn)(coakka_v2_runtime_t *, coakka_v2_host_handles_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_start_fn)(coakka_v2_runtime_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_stop_fn)(coakka_v2_runtime_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_health_fn)(coakka_v2_runtime_t *, coakka_v2_runtime_health_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_stats_fn)(coakka_v2_runtime_t *, coakka_v2_runtime_stats_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_submit_envelope_fn)(coakka_v2_runtime_t *, const uint8_t *, size_t);
typedef coakka_v2_status_t (*coakka_v2_runtime_apply_control_envelope_fn)(coakka_v2_runtime_t *, const uint8_t *, size_t);
typedef coakka_v2_status_t (*coakka_v2_monitor_wait_consume_fn)(int, uint32_t, uint64_t *);
typedef void (*coakka_v2_host_handles_close_fn)(coakka_v2_host_handles_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_request_reader_create_fn)(coakka_v2_runtime_t *, const coakka_v2_host_handles_t *, coakka_v2_request_reader_t **);
typedef coakka_v2_status_t (*coakka_v2_request_reader_read_fn)(coakka_v2_request_reader_t *, uint32_t, coakka_v2_owned_frame_t **);
typedef const uint8_t *(*coakka_v2_owned_frame_data_fn)(const coakka_v2_owned_frame_t *);
typedef size_t (*coakka_v2_owned_frame_size_fn)(const coakka_v2_owned_frame_t *);
typedef void (*coakka_v2_owned_frame_release_fn)(coakka_v2_owned_frame_t *);
typedef void (*coakka_v2_request_reader_destroy_fn)(coakka_v2_request_reader_t *);
typedef coakka_v2_status_t (*coakka_v2_ask_client_create_ex_fn)(coakka_v2_runtime_t *, const coakka_v2_host_handles_t *, coakka_v2_ask_client_t **);
typedef void (*coakka_v2_ask_client_destroy_fn)(coakka_v2_ask_client_t *);
typedef coakka_v2_status_t (*coakka_v2_ask_client_begin_fn)(coakka_v2_ask_client_t *, const uint8_t *, size_t, coakka_v2_ask_ticket_t **);
typedef coakka_v2_status_t (*coakka_v2_ask_ticket_await_fn)(coakka_v2_ask_ticket_t *, uint32_t, uint32_t *, uint8_t **, size_t *);
typedef void (*coakka_v2_ask_ticket_cancel_fn)(coakka_v2_ask_ticket_t *);
typedef void (*coakka_v2_ask_ticket_destroy_fn)(coakka_v2_ask_ticket_t *);
typedef coakka_v2_status_t (*coakka_v2_client_copy_request_dispatch_fn)(const uint8_t *, size_t, uint8_t **, size_t *, uint8_t **, size_t *, uint32_t *);
typedef coakka_v2_status_t (*coakka_v2_client_copy_envelope_payload_fn)(const uint8_t *, size_t, uint32_t, uint8_t **, size_t *);
typedef coakka_v2_status_t (*coakka_v2_client_build_raw_request_fn)(const coakka_v2_client_raw_request_spec_t *, uint8_t **, size_t *);
typedef coakka_v2_status_t (*coakka_v2_client_build_raw_reply_fn)(const coakka_v2_client_raw_reply_spec_t *, uint8_t **, size_t *);
typedef void (*coakka_v2_client_bytes_release_fn)(uint8_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_apply_tcp_connection_options_ex_fn)(coakka_v2_runtime_t *, const coakka_v2_tcp_connection_options_t *, coakka_v2_tcp_connection_apply_result_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_tcp_connection_config_fn)(coakka_v2_runtime_t *, coakka_v2_tcp_connection_config_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_apply_tcp_security_options_ex_fn)(coakka_v2_runtime_t *, const coakka_v2_tcp_security_options_t *, coakka_v2_tcp_security_apply_result_t *);
typedef coakka_v2_status_t (*coakka_v2_runtime_get_tcp_security_info_fn)(coakka_v2_runtime_t *, coakka_v2_tcp_security_info_t *);
typedef const char *(*coakka_v2_transport_apply_reason_name_fn)(uint32_t);
typedef coakka_v2_status_t (*coakka_v2_file_lane_create_ex_fn)(const coakka_v2_file_lane_config_t *, coakka_v2_file_lane_t **);
typedef void (*coakka_v2_file_lane_destroy_fn)(coakka_v2_file_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_start_fn)(coakka_v2_file_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_stop_fn)(coakka_v2_file_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_get_bound_port_fn)(coakka_v2_file_lane_t *, uint16_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_prepare_receive_fn)(coakka_v2_file_lane_t *, const coakka_v2_file_receive_spec_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_submit_send_fn)(coakka_v2_file_lane_t *, const coakka_v2_file_send_spec_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_get_transfer_fn)(coakka_v2_file_lane_t *, const char *, uint32_t, coakka_v2_file_transfer_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_wait_transfer_fn)(coakka_v2_file_lane_t *, const char *, uint32_t, uint64_t, uint32_t, coakka_v2_file_transfer_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_cancel_transfer_fn)(coakka_v2_file_lane_t *, const char *, uint32_t);
typedef coakka_v2_status_t (*coakka_v2_file_lane_forget_transfer_fn)(coakka_v2_file_lane_t *, const char *, uint32_t);
typedef coakka_v2_status_t (*coakka_v2_file_lane_get_stats_fn)(coakka_v2_file_lane_t *, coakka_v2_file_lane_stats_t *);
typedef coakka_v2_status_t (*coakka_v2_file_sha256_path_fn)(const char *, uint8_t *, uint64_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_create_ex_fn)(const coakka_v2_stream_lane_config_t *, coakka_v2_stream_lane_t **);
typedef void (*coakka_v2_stream_lane_destroy_fn)(coakka_v2_stream_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_start_fn)(coakka_v2_stream_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_stop_fn)(coakka_v2_stream_lane_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_get_bound_port_fn)(coakka_v2_stream_lane_t *, uint16_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_prepare_publish_fn)(coakka_v2_stream_lane_t *, const coakka_v2_stream_publish_spec_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_subscribe_fn)(coakka_v2_stream_lane_t *, const coakka_v2_stream_subscribe_spec_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_get_session_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t, coakka_v2_stream_session_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_wait_session_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t, uint64_t, uint32_t, coakka_v2_stream_session_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_get_pressure_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t, coakka_v2_stream_pressure_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_wait_pressure_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t, uint64_t, uint32_t, coakka_v2_stream_pressure_snapshot_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_cancel_session_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_forget_session_fn)(coakka_v2_stream_lane_t *, const char *, uint32_t);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_get_stats_fn)(coakka_v2_stream_lane_t *, coakka_v2_stream_lane_stats_t *);
typedef coakka_v2_status_t (*coakka_v2_file_lane_create_owned_ex_fn)(const coakka_v2_file_lane_owned_config_t *, coakka_v2_file_lane_t **);
typedef coakka_v2_status_t (*coakka_v2_file_lane_prepare_receive_grant_fn)(coakka_v2_file_lane_t *, const coakka_v2_file_receive_spec_t *, coakka_v2_file_receive_grant_t *);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_create_owned_ex_fn)(const coakka_v2_stream_lane_owned_config_t *, coakka_v2_stream_lane_t **);
typedef coakka_v2_status_t (*coakka_v2_stream_lane_prepare_publish_grant_fn)(coakka_v2_stream_lane_t *, const coakka_v2_stream_publish_spec_t *, coakka_v2_stream_publish_grant_t *);

struct coakka_v2_go_bindings_t {
  void *handle;
  coakka_v2_runtime_get_abi_version_fn get_abi_version;
  coakka_v2_runtime_get_info_fn get_info;
  coakka_v2_runtime_get_config_fn get_config;
  coakka_v2_runtime_get_capabilities_fn get_capabilities;
  coakka_v2_runtime_create_fn create_runtime;
  coakka_v2_runtime_create_native_fn create_runtime_native;
  coakka_v2_runtime_get_startup_config_fn get_startup_config;
  coakka_v2_runtime_destroy_fn destroy_runtime;
  coakka_v2_runtime_apply_network_options_fn apply_network_options;
  coakka_v2_runtime_get_network_config_fn get_network_config;
  coakka_v2_client_encode_message_value_fn encode_message_value;
  coakka_v2_client_decode_message_value_fn decode_message_value;
  coakka_v2_client_decode_deadletter_value_fn decode_deadletter_value;
  coakka_v2_message_value_release_fn message_value_release;
  coakka_v2_client_encode_control_value_fn encode_control_value;
  coakka_v2_runtime_get_host_handles_fn get_host_handles;
  coakka_v2_runtime_start_fn start_runtime;
  coakka_v2_runtime_stop_fn stop_runtime;
  coakka_v2_runtime_get_health_fn get_health;
  coakka_v2_runtime_get_stats_fn get_stats;
  coakka_v2_runtime_submit_envelope_fn submit_envelope;
  coakka_v2_runtime_apply_control_envelope_fn apply_control_envelope;
  coakka_v2_monitor_wait_consume_fn monitor_wait_consume;
  coakka_v2_host_handles_close_fn host_handles_close;
  coakka_v2_runtime_request_reader_create_fn request_reader_create;
  coakka_v2_request_reader_read_fn request_reader_read;
  coakka_v2_owned_frame_data_fn owned_frame_data;
  coakka_v2_owned_frame_size_fn owned_frame_size;
  coakka_v2_owned_frame_release_fn owned_frame_release;
  coakka_v2_request_reader_destroy_fn request_reader_destroy;
  coakka_v2_ask_client_create_ex_fn ask_client_create_ex;
  coakka_v2_ask_client_destroy_fn ask_client_destroy;
  coakka_v2_ask_client_begin_fn ask_client_begin;
  coakka_v2_ask_ticket_await_fn ask_ticket_await;
  coakka_v2_ask_ticket_cancel_fn ask_ticket_cancel;
  coakka_v2_ask_ticket_destroy_fn ask_ticket_destroy;
  coakka_v2_client_copy_request_dispatch_fn copy_request_dispatch;
  coakka_v2_client_copy_envelope_payload_fn copy_envelope_payload;
  coakka_v2_client_build_raw_request_fn build_raw_request;
  coakka_v2_client_build_raw_reply_fn build_raw_reply;
  coakka_v2_client_bytes_release_fn client_bytes_release;
  coakka_v2_runtime_apply_tcp_connection_options_ex_fn apply_tcp_connection_options_ex;
  coakka_v2_runtime_get_tcp_connection_config_fn get_tcp_connection_config;
  coakka_v2_runtime_apply_tcp_security_options_ex_fn apply_tcp_security_options_ex;
  coakka_v2_runtime_get_tcp_security_info_fn get_tcp_security_info;
  coakka_v2_transport_apply_reason_name_fn transport_apply_reason_name;
  coakka_v2_file_lane_create_ex_fn file_lane_create_ex;
  coakka_v2_file_lane_create_owned_ex_fn file_lane_create_owned_ex;
  coakka_v2_file_lane_destroy_fn file_lane_destroy;
  coakka_v2_file_lane_start_fn file_lane_start;
  coakka_v2_file_lane_stop_fn file_lane_stop;
  coakka_v2_file_lane_get_bound_port_fn file_lane_get_bound_port;
  coakka_v2_file_lane_prepare_receive_fn file_lane_prepare_receive;
  coakka_v2_file_lane_prepare_receive_grant_fn file_lane_prepare_receive_grant;
  coakka_v2_file_lane_submit_send_fn file_lane_submit_send;
  coakka_v2_file_lane_get_transfer_fn file_lane_get_transfer;
  coakka_v2_file_lane_wait_transfer_fn file_lane_wait_transfer;
  coakka_v2_file_lane_cancel_transfer_fn file_lane_cancel_transfer;
  coakka_v2_file_lane_forget_transfer_fn file_lane_forget_transfer;
  coakka_v2_file_lane_get_stats_fn file_lane_get_stats;
  coakka_v2_file_sha256_path_fn file_sha256_path;
  coakka_v2_stream_lane_create_ex_fn stream_lane_create_ex;
  coakka_v2_stream_lane_create_owned_ex_fn stream_lane_create_owned_ex;
  coakka_v2_stream_lane_destroy_fn stream_lane_destroy;
  coakka_v2_stream_lane_start_fn stream_lane_start;
  coakka_v2_stream_lane_stop_fn stream_lane_stop;
  coakka_v2_stream_lane_get_bound_port_fn stream_lane_get_bound_port;
  coakka_v2_stream_lane_prepare_publish_fn stream_lane_prepare_publish;
  coakka_v2_stream_lane_prepare_publish_grant_fn stream_lane_prepare_publish_grant;
  coakka_v2_stream_lane_subscribe_fn stream_lane_subscribe;
  coakka_v2_stream_lane_get_session_fn stream_lane_get_session;
  coakka_v2_stream_lane_wait_session_fn stream_lane_wait_session;
  coakka_v2_stream_lane_get_pressure_fn stream_lane_get_pressure;
  coakka_v2_stream_lane_wait_pressure_fn stream_lane_wait_pressure;
  coakka_v2_stream_lane_cancel_session_fn stream_lane_cancel_session;
  coakka_v2_stream_lane_forget_session_fn stream_lane_forget_session;
  coakka_v2_stream_lane_get_stats_fn stream_lane_get_stats;
};

static void *coakka_v2_go_process_library_handle = NULL;
static char *coakka_v2_go_process_library_path = NULL;
/* Open is cold/serialized by the Go module gate. Only per-binding tables are
 * freed; the one admitted module/path remains process-owned for native TLS. */

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

extern int coakka_v2_go_stream_source_next(uintptr_t context,
                                            uint8_t *destination,
                                            size_t capacity,
                                            coakka_v2_stream_frame_t *out_frame);
extern int coakka_v2_go_stream_consume(uintptr_t context, uint8_t *data,
                                       coakka_v2_stream_frame_t *frame);

static coakka_v2_status_t coakka_v2_go_stream_source_trampoline(
    void *context, uint8_t *destination, size_t capacity,
    coakka_v2_stream_frame_t *out_frame) {
  return (coakka_v2_status_t)coakka_v2_go_stream_source_next((uintptr_t)context, destination,
                                          capacity, out_frame);
}

static coakka_v2_status_t coakka_v2_go_stream_consumer_trampoline(
    void *context, const uint8_t *data,
    const coakka_v2_stream_frame_t *frame) {
  /* cgo has no const pointer spelling. Go only borrows these
   * inputs for read-only projection; neither callback may mutate/retain them. */
  return (coakka_v2_status_t)coakka_v2_go_stream_consume((uintptr_t)context,
      (uint8_t *)data, (coakka_v2_stream_frame_t *)frame);
}

_Static_assert(_Generic(&coakka_v2_go_stream_source_trampoline, coakka_v2_stream_source_next_fn: 1, default: 0), "Core source callback mismatch");
_Static_assert(_Generic(&coakka_v2_go_stream_consumer_trampoline, coakka_v2_stream_consumer_fn: 1, default: 0), "Core consumer callback mismatch");

#ifdef _WIN32
static char *coakka_v2_go_windows_error_code(const char *operation, DWORD code) {
  char message[128];
  snprintf(message, sizeof(message), "%s failed win32_error=%lu", operation,
           (unsigned long)code);
  return coakka_v2_go_duplicate_text(message);
}

static char *coakka_v2_go_windows_error(const char *operation) {
  return coakka_v2_go_windows_error_code(operation, GetLastError());
}

static void *coakka_v2_go_windows_load_module(const char *path, char **error_out) {
  size_t bytes = 0u;
  DWORD error = ERROR_SUCCESS;
  const char *operation = "MultiByteToWideChar";
  WCHAR *wide = NULL;
  HMODULE module = NULL;
  if (path == NULL || path[0] == '\0') {
    error = ERROR_INVALID_PARAMETER;
    goto failed;
  }
  while (bytes <= COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_BYTES &&
         path[bytes] != '\0') ++bytes;
  if (bytes > COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_BYTES) {
    error = ERROR_FILENAME_EXCED_RANGE;
    goto failed;
  }
  /* Bound bytes before narrowing and query units before allocating. Strict
   * UTF-8 never silently selects another ANSI name or replacement character. */
  const int units = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS,
      path, (int)bytes + 1, NULL, 0);
  if (units == 0) {
    error = GetLastError();
    goto failed;
  }
  if (units > COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_UNITS) {
    error = ERROR_FILENAME_EXCED_RANGE;
    goto failed;
  }
  wide = (WCHAR *)malloc((size_t)units * sizeof(*wide));
  if (wide == NULL) {
    operation = "runtime path allocation";
    error = ERROR_NOT_ENOUGH_MEMORY;
    goto failed;
  }
  const int converted = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS,
      path, (int)bytes + 1, wide, units);
  if (converted != units) {
    error = converted == 0 ? GetLastError() : ERROR_INVALID_DATA;
    goto failed;
  }
  operation = "LoadLibraryW";
  module = LoadLibraryW(wide);
  if (module == NULL) error = GetLastError();
  /* Cleanup may change last-error. Keep the causal operation/code, never
   * retry under another encoding or mutate process-wide search policy. */
  free(wide);
  wide = NULL;
  if (module != NULL) return (void *)module;
failed:
  free(wide);
  if (error_out != NULL) {
    *error_out = coakka_v2_go_windows_error_code(operation, error);
  }
  return NULL;
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

/* Function-pointer slots are never written through an aliased void**.
 * C11 proofs bind each load to the owning SDK declaration at compile time. */
static int coakka_v2_go_load_typed_symbol(void *handle, void *slot, size_t width,
    const char *name, char **error_out) {
  void *symbol = NULL;
  int status = coakka_v2_go_load_symbol(handle, &symbol, name, error_out);
  if (status == 0) memcpy(slot, &symbol, width);
  return status;
}
#define COAKKA_GO_LOAD(owner, field, type, symbol, error) \
  ((void)sizeof(char[_Generic(&symbol, type: 1, default: -1)]), \
   (void)sizeof(char[_Generic(&(owner)->field, type *: 1, default: -1)]), \
   (void)sizeof(char[sizeof((owner)->field) == sizeof(void *) ? 1 : -1]), \
   coakka_v2_go_load_typed_symbol((owner)->handle, &(owner)->field, \
       sizeof((owner)->field), #symbol, error))

coakka_v2_go_bindings_t *coakka_v2_go_open_library(const char *path, char **error_out) {
  int opened_new_library = 0;
  void *handle = NULL;
  if (error_out != NULL) *error_out = NULL;
  if (path == NULL) {
    if (error_out != NULL) {
      *error_out = coakka_v2_go_duplicate_text("runtime library path is null");
    }
    return NULL;
  }
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
    handle = coakka_v2_go_windows_load_module(path, error_out);
#else
    handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
#endif
    if (handle == NULL) {
#ifndef _WIN32
      if (error_out != NULL) {
        const char *error = dlerror();
        *error_out = error == NULL
            ? coakka_v2_go_duplicate_text("dlopen failed")
            : coakka_v2_go_duplicate_text(error);
      }
#endif
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
  if (COAKKA_GO_LOAD(bindings, get_abi_version, coakka_v2_runtime_get_abi_version_fn, coakka_v2_runtime_get_abi_version, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_info, coakka_v2_runtime_get_info_fn, coakka_v2_runtime_get_info, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_config, coakka_v2_runtime_get_config_fn, coakka_v2_runtime_get_config, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_capabilities, coakka_v2_runtime_get_capabilities_fn, coakka_v2_runtime_get_capabilities, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, create_runtime, coakka_v2_runtime_create_fn, coakka_v2_runtime_create, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, create_runtime_native, coakka_v2_runtime_create_native_fn, coakka_v2_runtime_create_native, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_startup_config, coakka_v2_runtime_get_startup_config_fn, coakka_v2_runtime_get_startup_config, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, destroy_runtime, coakka_v2_runtime_destroy_fn, coakka_v2_runtime_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, apply_network_options, coakka_v2_runtime_apply_network_options_fn, coakka_v2_runtime_apply_network_options, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_network_config, coakka_v2_runtime_get_network_config_fn, coakka_v2_runtime_get_network_config, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, encode_message_value, coakka_v2_client_encode_message_value_fn, coakka_v2_client_encode_message_value, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, decode_message_value, coakka_v2_client_decode_message_value_fn, coakka_v2_client_decode_message_value, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, decode_deadletter_value, coakka_v2_client_decode_deadletter_value_fn, coakka_v2_client_decode_deadletter_value, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, message_value_release, coakka_v2_message_value_release_fn, coakka_v2_message_value_release, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, encode_control_value, coakka_v2_client_encode_control_value_fn, coakka_v2_client_encode_control_value, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_host_handles, coakka_v2_runtime_get_host_handles_fn, coakka_v2_runtime_get_host_handles, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, start_runtime, coakka_v2_runtime_start_fn, coakka_v2_runtime_start, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stop_runtime, coakka_v2_runtime_stop_fn, coakka_v2_runtime_stop, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_health, coakka_v2_runtime_get_health_fn, coakka_v2_runtime_get_health, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_stats, coakka_v2_runtime_get_stats_fn, coakka_v2_runtime_get_stats, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, submit_envelope, coakka_v2_runtime_submit_envelope_fn, coakka_v2_runtime_submit_envelope, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, apply_control_envelope, coakka_v2_runtime_apply_control_envelope_fn, coakka_v2_runtime_apply_control_envelope, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, monitor_wait_consume, coakka_v2_monitor_wait_consume_fn, coakka_v2_monitor_wait_consume, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, host_handles_close, coakka_v2_host_handles_close_fn, coakka_v2_host_handles_close, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, request_reader_create, coakka_v2_runtime_request_reader_create_fn, coakka_v2_runtime_request_reader_create, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, request_reader_read, coakka_v2_request_reader_read_fn, coakka_v2_request_reader_read, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, owned_frame_data, coakka_v2_owned_frame_data_fn, coakka_v2_owned_frame_data, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, owned_frame_size, coakka_v2_owned_frame_size_fn, coakka_v2_owned_frame_size, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, owned_frame_release, coakka_v2_owned_frame_release_fn, coakka_v2_owned_frame_release, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, request_reader_destroy, coakka_v2_request_reader_destroy_fn, coakka_v2_request_reader_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_client_create_ex, coakka_v2_ask_client_create_ex_fn, coakka_v2_ask_client_create_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_client_destroy, coakka_v2_ask_client_destroy_fn, coakka_v2_ask_client_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_client_begin, coakka_v2_ask_client_begin_fn, coakka_v2_ask_client_begin, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_ticket_await, coakka_v2_ask_ticket_await_fn, coakka_v2_ask_ticket_await, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_ticket_cancel, coakka_v2_ask_ticket_cancel_fn, coakka_v2_ask_ticket_cancel, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, ask_ticket_destroy, coakka_v2_ask_ticket_destroy_fn, coakka_v2_ask_ticket_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, copy_request_dispatch, coakka_v2_client_copy_request_dispatch_fn, coakka_v2_client_copy_request_dispatch, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, copy_envelope_payload, coakka_v2_client_copy_envelope_payload_fn, coakka_v2_client_copy_envelope_payload, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, build_raw_request, coakka_v2_client_build_raw_request_fn, coakka_v2_client_build_raw_request, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, build_raw_reply, coakka_v2_client_build_raw_reply_fn, coakka_v2_client_build_raw_reply, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, client_bytes_release, coakka_v2_client_bytes_release_fn, coakka_v2_client_bytes_release, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, apply_tcp_connection_options_ex, coakka_v2_runtime_apply_tcp_connection_options_ex_fn, coakka_v2_runtime_apply_tcp_connection_options_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_tcp_connection_config, coakka_v2_runtime_get_tcp_connection_config_fn, coakka_v2_runtime_get_tcp_connection_config, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, apply_tcp_security_options_ex, coakka_v2_runtime_apply_tcp_security_options_ex_fn, coakka_v2_runtime_apply_tcp_security_options_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, get_tcp_security_info, coakka_v2_runtime_get_tcp_security_info_fn, coakka_v2_runtime_get_tcp_security_info, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, transport_apply_reason_name, coakka_v2_transport_apply_reason_name_fn, coakka_v2_transport_apply_reason_name, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_create_ex, coakka_v2_file_lane_create_ex_fn, coakka_v2_file_lane_create_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_destroy, coakka_v2_file_lane_destroy_fn, coakka_v2_file_lane_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_start, coakka_v2_file_lane_start_fn, coakka_v2_file_lane_start, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_stop, coakka_v2_file_lane_stop_fn, coakka_v2_file_lane_stop, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_get_bound_port, coakka_v2_file_lane_get_bound_port_fn, coakka_v2_file_lane_get_bound_port, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_prepare_receive, coakka_v2_file_lane_prepare_receive_fn, coakka_v2_file_lane_prepare_receive, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_submit_send, coakka_v2_file_lane_submit_send_fn, coakka_v2_file_lane_submit_send, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_get_transfer, coakka_v2_file_lane_get_transfer_fn, coakka_v2_file_lane_get_transfer, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_wait_transfer, coakka_v2_file_lane_wait_transfer_fn, coakka_v2_file_lane_wait_transfer, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_cancel_transfer, coakka_v2_file_lane_cancel_transfer_fn, coakka_v2_file_lane_cancel_transfer, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_forget_transfer, coakka_v2_file_lane_forget_transfer_fn, coakka_v2_file_lane_forget_transfer, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_get_stats, coakka_v2_file_lane_get_stats_fn, coakka_v2_file_lane_get_stats, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_sha256_path, coakka_v2_file_sha256_path_fn, coakka_v2_file_sha256_path, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_create_ex, coakka_v2_stream_lane_create_ex_fn, coakka_v2_stream_lane_create_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_destroy, coakka_v2_stream_lane_destroy_fn, coakka_v2_stream_lane_destroy, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_start, coakka_v2_stream_lane_start_fn, coakka_v2_stream_lane_start, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_stop, coakka_v2_stream_lane_stop_fn, coakka_v2_stream_lane_stop, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_get_bound_port, coakka_v2_stream_lane_get_bound_port_fn, coakka_v2_stream_lane_get_bound_port, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_prepare_publish, coakka_v2_stream_lane_prepare_publish_fn, coakka_v2_stream_lane_prepare_publish, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_subscribe, coakka_v2_stream_lane_subscribe_fn, coakka_v2_stream_lane_subscribe, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_get_session, coakka_v2_stream_lane_get_session_fn, coakka_v2_stream_lane_get_session, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_wait_session, coakka_v2_stream_lane_wait_session_fn, coakka_v2_stream_lane_wait_session, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_get_pressure, coakka_v2_stream_lane_get_pressure_fn, coakka_v2_stream_lane_get_pressure, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_wait_pressure, coakka_v2_stream_lane_wait_pressure_fn, coakka_v2_stream_lane_wait_pressure, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_cancel_session, coakka_v2_stream_lane_cancel_session_fn, coakka_v2_stream_lane_cancel_session, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_forget_session, coakka_v2_stream_lane_forget_session_fn, coakka_v2_stream_lane_forget_session, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_get_stats, coakka_v2_stream_lane_get_stats_fn, coakka_v2_stream_lane_get_stats, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_create_owned_ex, coakka_v2_file_lane_create_owned_ex_fn, coakka_v2_file_lane_create_owned_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, file_lane_prepare_receive_grant, coakka_v2_file_lane_prepare_receive_grant_fn, coakka_v2_file_lane_prepare_receive_grant, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_create_owned_ex, coakka_v2_stream_lane_create_owned_ex_fn, coakka_v2_stream_lane_create_owned_ex, error_out) != 0 ||
      COAKKA_GO_LOAD(bindings, stream_lane_prepare_publish_grant, coakka_v2_stream_lane_prepare_publish_grant_fn, coakka_v2_stream_lane_prepare_publish_grant, error_out) != 0) {
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

void coakka_v2_go_close_library(coakka_v2_go_bindings_t *bindings) {
  if (bindings == NULL) {
    return;
  }
  free(bindings);
}

uint32_t coakka_v2_go_get_abi_version(coakka_v2_go_bindings_t *bindings) {
  return bindings->get_abi_version();
}

int coakka_v2_go_get_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_info_t *out_info) {
  return bindings->get_info(out_info);
}

int coakka_v2_go_get_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_config_view_t *out_config) {
  return bindings->get_config(rt, out_config);
}

int coakka_v2_go_get_capabilities(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_capability_snapshot_t *out_capabilities) {
  return bindings->get_capabilities(out_capabilities);
}

int coakka_v2_go_create_runtime_native(coakka_v2_go_bindings_t *bindings,
                                               const coakka_v2_runtime_config_t *cfg,
                                               const coakka_v2_tcp_connection_options_t *connection,
                                               const coakka_v2_startup_monitor_options_t *monitor,
                                               coakka_v2_runtime_t **out_runtime) {
  coakka_v2_runtime_startup_options_t startup = {
    sizeof(coakka_v2_runtime_startup_options_t), NULL, connection, NULL, monitor
  };
  return bindings->create_runtime_native(cfg, &startup, out_runtime);
}

int coakka_v2_go_get_startup_config(coakka_v2_go_bindings_t *bindings,
                                            coakka_v2_runtime_t *runtime,
                                            coakka_v2_runtime_startup_config_t *out_config) {
  return bindings->get_startup_config(runtime, out_config);
}

void coakka_v2_go_destroy_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  bindings->destroy_runtime(rt);
}

int coakka_v2_go_get_network_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_network_config_snapshot_t *out) {
  return bindings->get_network_config(rt, out);
}
int coakka_v2_go_encode_message_value(coakka_v2_go_bindings_t *bindings, const coakka_v2_message_value_t *value, uint8_t **out, size_t *size) {
  return bindings->encode_message_value(value, out, size);
}
int coakka_v2_go_decode_message_value(coakka_v2_go_bindings_t *bindings, const uint8_t *data, size_t size, coakka_v2_message_value_owner_t **owner, coakka_v2_message_value_t *out) {
  return bindings->decode_message_value(data, size, owner, out);
}
int coakka_v2_go_decode_deadletter_value(coakka_v2_go_bindings_t *bindings, const uint8_t *data, size_t size, coakka_v2_message_value_owner_t **owner, coakka_v2_deadletter_value_t *out) {
  return bindings->decode_deadletter_value(data, size, owner, out);
}
void coakka_v2_go_message_value_release(coakka_v2_go_bindings_t *bindings, coakka_v2_message_value_owner_t *owner) {
  bindings->message_value_release(owner);
}
int coakka_v2_go_encode_control_value(coakka_v2_go_bindings_t *bindings, const coakka_v2_control_value_t *value, uint8_t **out, size_t *size) {
  return bindings->encode_control_value(value, out, size);
}

int coakka_v2_go_apply_network_options(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_network_options_t *options) {
  return bindings->apply_network_options(rt, options);
}

int coakka_v2_go_get_host_handles(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_host_handles_t *out_handles) {
  return bindings->get_host_handles(rt, out_handles);
}

int coakka_v2_go_start_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  return bindings->start_runtime(rt);
}

int coakka_v2_go_stop_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt) {
  return bindings->stop_runtime(rt);
}

int coakka_v2_go_get_health(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_health_t *out_health) {
  return bindings->get_health(rt, out_health);
}

int coakka_v2_go_get_stats(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_stats_t *out_stats) {
  return bindings->get_stats(rt, out_stats);
}

int coakka_v2_go_submit_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len) {
  return bindings->submit_envelope(rt, buf, len);
}

int coakka_v2_go_apply_control_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len) {
  return bindings->apply_control_envelope(rt, buf, len);
}

int coakka_v2_go_monitor_wait_consume(coakka_v2_go_bindings_t *bindings,
                                              int fd, uint32_t timeout_ms,
                                              uint64_t *out_signal_count) {
  return bindings->monitor_wait_consume(fd, timeout_ms, out_signal_count);
}

void coakka_v2_go_host_handles_close(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_host_handles_t *handles) {
  bindings->host_handles_close(handles);
}

int coakka_v2_go_request_reader_create(coakka_v2_go_bindings_t *bindings,
                                               coakka_v2_runtime_t *runtime,
                                               const coakka_v2_host_handles_t *handles,
                                               coakka_v2_request_reader_t **out_reader) {
  return bindings->request_reader_create(runtime, handles, out_reader);
}

void coakka_v2_go_request_reader_destroy(coakka_v2_go_bindings_t *bindings,
                                                 coakka_v2_request_reader_t *reader) {
  bindings->request_reader_destroy(reader);
}

int coakka_v2_go_ask_client_create_ex(
    coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *runtime,
    const coakka_v2_host_handles_t *handles, coakka_v2_ask_client_t **out_client) {
  return bindings->ask_client_create_ex(runtime, handles, out_client);
}

void coakka_v2_go_ask_client_destroy(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_ask_client_t *client) {
  bindings->ask_client_destroy(client);
}

int coakka_v2_go_ask_client_begin(coakka_v2_go_bindings_t *bindings,
                                          coakka_v2_ask_client_t *client,
                                          const uint8_t *frame, size_t frame_len,
                                          coakka_v2_ask_ticket_t **out_ticket) {
  return bindings->ask_client_begin(client, frame, frame_len, out_ticket);
}

int coakka_v2_go_ask_ticket_await(coakka_v2_go_bindings_t *bindings,
                                          coakka_v2_ask_ticket_t *ticket,
                                          uint32_t timeout_ms, uint32_t *out_kind,
                                          uint8_t **out_frame, size_t *out_frame_len) {
  return bindings->ask_ticket_await(ticket, timeout_ms, out_kind, out_frame,
                                    out_frame_len);
}

void coakka_v2_go_ask_ticket_cancel(coakka_v2_go_bindings_t *bindings,
                                            coakka_v2_ask_ticket_t *ticket) {
  bindings->ask_ticket_cancel(ticket);
}

void coakka_v2_go_ask_ticket_destroy(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_ask_ticket_t *ticket) {
  bindings->ask_ticket_destroy(ticket);
}

// The common request fits one reader-owned Go scratch region. Core's owners
// are released before return and C retains no Go pointer. Oversized records use
// one bounded C allocation so the public maximum remains unchanged.
static int coakka_v2_go_pack_delivered_request(
    coakka_v2_go_bindings_t *bindings, coakka_v2_request_reader_t *reader,
    uint32_t timeout_ms, uint8_t *staging, size_t staging_capacity,
    coakka_v2_go_packed_request_t *out_request) {
  coakka_v2_owned_frame_t *frame_owner = NULL;
  uint8_t *target = NULL;
  uint8_t *payload = NULL;
  const uint8_t *frame = NULL;
  size_t frame_len = 0U;
  size_t target_len = 0U;
  size_t payload_len = 0U;
  uint32_t one_way = 0U;
  uint8_t *destination = NULL;
  int status;

  if (staging == NULL || staging_capacity == 0U || out_request == NULL) {
    return -1;
  }
  memset(out_request, 0, sizeof(*out_request));
  status = bindings->request_reader_read(reader, timeout_ms, &frame_owner);
  if (status != 0) {
    return status;
  }
  if (frame_owner == NULL) {
    return -5;
  }
  frame = bindings->owned_frame_data(frame_owner);
  frame_len = bindings->owned_frame_size(frame_owner);
  status = bindings->copy_request_dispatch(
      frame, frame_len, &target, &target_len, &payload, &payload_len,
      &one_way);
  if (status != 0) {
    goto cleanup;
  }
  if (frame == NULL || frame_len == 0U || target == NULL || target_len == 0U ||
      (payload_len != 0U && payload == NULL) || one_way > 1U ||
      frame_len > SIZE_MAX - target_len ||
      frame_len + target_len == SIZE_MAX ||
      frame_len + target_len + 1U > SIZE_MAX - payload_len) {
    status = -5;
    goto cleanup;
  }
  out_request->packed_len = frame_len + target_len + 1U + payload_len;
  if (out_request->packed_len <= staging_capacity) {
    destination = staging;
  } else {
    out_request->large_packed = (uint8_t *)malloc(out_request->packed_len);
    if (out_request->large_packed == NULL) {
      status = -2;
      goto cleanup;
    }
    destination = out_request->large_packed;
  }
  memcpy(destination, frame, frame_len);
  memcpy(destination + frame_len, target, target_len);
  destination[frame_len + target_len] = 0U;
  if (payload_len != 0U) {
    memcpy(destination + frame_len + target_len + 1U, payload, payload_len);
  }
  out_request->frame_len = frame_len;
  out_request->target_len = target_len;
  out_request->payload_len = payload_len;
  out_request->one_way = one_way;

cleanup:
  if (target != NULL) {
    bindings->client_bytes_release(target);
  }
  if (payload != NULL) {
    bindings->client_bytes_release(payload);
  }
  bindings->owned_frame_release(frame_owner);
  if (status != 0 && out_request->large_packed != NULL) {
    free(out_request->large_packed);
    memset(out_request, 0, sizeof(*out_request));
  }
  return status;
}

coakka_v2_go_packed_result_t coakka_v2_go_read_delivered_request_packed(
    coakka_v2_go_bindings_t *bindings, coakka_v2_request_reader_t *reader,
    uint32_t timeout_ms, uint8_t *staging, size_t staging_capacity) {
  coakka_v2_go_packed_result_t result = {0};
  result.status = coakka_v2_go_pack_delivered_request(
      bindings, reader, timeout_ms, staging, staging_capacity, &result.value);
  return result;
}

// Response asks return only the final payload to Go; deadletters retain their
// complete frame for the compatibility value projection. Core still builds,
// validates and classifies both results.
static int coakka_v2_go_ask_ticket_await_payload(
    coakka_v2_go_bindings_t *bindings, coakka_v2_ask_ticket_t *ticket,
    uint32_t timeout_ms, uint32_t *out_kind, uint8_t **out_bytes,
    size_t *out_bytes_len) {
  uint8_t *frame = NULL;
  size_t frame_len = 0U;
  if (out_kind == NULL || out_bytes == NULL || out_bytes_len == NULL) {
    return -1;
  }
  *out_kind = 0U;
  *out_bytes = NULL;
  *out_bytes_len = 0U;
  int status = bindings->ask_ticket_await(ticket, timeout_ms, out_kind, &frame,
                                           &frame_len);
  if (status != 0) {
    if (frame != NULL) {
      bindings->client_bytes_release(frame);
    }
    return status;
  }
  if (*out_kind == 1U) {
    status = bindings->copy_envelope_payload(frame, frame_len, 2U, out_bytes,
                                              out_bytes_len);
    bindings->client_bytes_release(frame);
    return status;
  }
  if (*out_kind == 2U) {
    *out_bytes = frame;
    *out_bytes_len = frame_len;
    return 0;
  }
  if (frame != NULL) {
    bindings->client_bytes_release(frame);
  }
  return -5;
}

static int coakka_v2_go_ask_client_begin_raw(
    coakka_v2_go_bindings_t *bindings, coakka_v2_ask_client_t *client,
    const char *message_id, const char *source, const char *target,
    const char *reply_to, const uint8_t *payload, size_t payload_len,
    uint32_t timeout_ms, uint32_t delivery_hint, uint32_t one_way,
    coakka_v2_ask_ticket_t **out_ticket) {
  uint8_t *frame = NULL;
  size_t frame_len = 0U;
  const coakka_v2_client_raw_request_spec_t spec = {
    sizeof(coakka_v2_client_raw_request_spec_t), message_id, source, target,
    reply_to, payload, payload_len, timeout_ms, delivery_hint, one_way
  };
  int status = bindings->build_raw_request(&spec, &frame, &frame_len);
  if (status == 0) {
    status = bindings->ask_client_begin(client, frame, frame_len, out_ticket);
  }
  if (frame != NULL) {
    bindings->client_bytes_release(frame);
  }
  return status;
}

coakka_v2_go_payload_result_t coakka_v2_go_ask_payload(
    coakka_v2_go_bindings_t *bindings, coakka_v2_ask_client_t *client,
    const char *message_id, const char *source, const char *target,
    const char *reply_to, const uint8_t *payload, size_t payload_len,
    uint32_t request_timeout_ms, uint32_t wait_timeout_ms,
    uint32_t delivery_hint) {
  coakka_v2_go_payload_result_t result = {0};
  coakka_v2_ask_ticket_t *ticket = NULL;
  uint8_t *bytes = NULL;
  size_t size = 0U;
  int status = coakka_v2_go_ask_client_begin_raw(
      bindings, client, message_id, source, target, reply_to, payload,
      payload_len, request_timeout_ms, delivery_hint, 0U, &ticket);
  if (status == 0) {
    status = coakka_v2_go_ask_ticket_await_payload(
        bindings, ticket, wait_timeout_ms, &result.kind, &bytes, &size);
  }
  if (ticket != NULL) {
    bindings->ask_ticket_destroy(ticket);
  }
  result.status = status;
  if (status != 0) {
    /* Partial projection owners retire before publishing the causal status. */
    if (bytes != NULL) bindings->client_bytes_release(bytes);
    return result;
  }
  result.value.size = size;
  if (size <= COAKKA_GO_INLINE_PAYLOAD_BYTES &&
      (size == 0U || bytes != NULL)) {
    if (size != 0U) memcpy(result.value.inline_bytes, bytes, size);
    if (bytes != NULL) bindings->client_bytes_release(bytes);
    result.value.inline_owned = 1U;
  } else {
    /* Oversized or malformed spans retain their exact extent/pointer for Go's
     * existing boundary checks. Only a real non-null byte owner is released. */
    result.value.owner = bytes;
  }
  return result;
}

int coakka_v2_go_submit_raw_reply(
    coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *runtime,
    const uint8_t *request_frame, size_t request_frame_len,
    const char *source, const uint8_t *payload, size_t payload_len) {
  uint8_t *reply = NULL;
  size_t reply_len = 0U;
  const coakka_v2_client_raw_reply_spec_t spec = {
    sizeof(coakka_v2_client_raw_reply_spec_t), request_frame,
    request_frame_len, source, payload, payload_len
  };
  int status = bindings->build_raw_reply(&spec, &reply, &reply_len);
  if (status == 0) {
    status = bindings->submit_envelope(runtime, reply, reply_len);
  }
  if (reply != NULL) {
    bindings->client_bytes_release(reply);
  }
  return status;
}

void coakka_v2_go_client_bytes_release(coakka_v2_go_bindings_t *bindings,
                                               uint8_t *bytes) {
  bindings->client_bytes_release(bytes);
}

int coakka_v2_go_apply_tcp_connection_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_connection_options_t *options, coakka_v2_tcp_connection_apply_result_t *out_result) {
  return bindings->apply_tcp_connection_options_ex(rt, options, out_result);
}

int coakka_v2_go_get_tcp_connection_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_connection_config_snapshot_t *out_config) {
  return bindings->get_tcp_connection_config(rt, out_config);
}

int coakka_v2_go_apply_tcp_security_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_security_options_t *options, coakka_v2_tcp_security_apply_result_t *out_result) {
  return bindings->apply_tcp_security_options_ex(rt, options, out_result);
}

int coakka_v2_go_get_tcp_security_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_security_info_t *out_info) {
  return bindings->get_tcp_security_info(rt, out_info);
}

const char *coakka_v2_go_transport_apply_reason_name(coakka_v2_go_bindings_t *bindings, uint32_t reason) {
  return bindings->transport_apply_reason_name(reason);
}

/* A loader table is ABI admission, never a capability probe. */
int coakka_v2_go_query_feature(coakka_v2_go_bindings_t *bindings,
    uint32_t feature, uint32_t *out_supported) {
  if (bindings == NULL || out_supported == NULL) return COAKKA_V2_ERR_INVALID_ARG;
  *out_supported = 0u;
  coakka_v2_runtime_info_t info = {0};
  info.struct_size = sizeof(info);
  int status = bindings->get_info(&info);
  if (status != COAKKA_V2_OK) return status;
  *out_supported = (info.feature_flags & feature) == feature ? 1u : 0u;
  return COAKKA_V2_OK;
}
int coakka_v2_go_file_lane_create_ex(coakka_v2_go_bindings_t *b, const coakka_v2_file_lane_config_t *c, coakka_v2_file_lane_t **out) { return b->file_lane_create_ex(c, out); }
int coakka_v2_go_file_lane_create_owned_ex(coakka_v2_go_bindings_t *b, const coakka_v2_file_lane_owned_config_t *c, coakka_v2_file_lane_t **out) { return b->file_lane_create_owned_ex(c, out); }
void coakka_v2_go_file_lane_destroy(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane) { b->file_lane_destroy(lane); }
int coakka_v2_go_file_lane_start(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane) { return b->file_lane_start(lane); }
int coakka_v2_go_file_lane_stop(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane) { return b->file_lane_stop(lane); }
int coakka_v2_go_file_lane_get_bound_port(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, uint16_t *out) { return b->file_lane_get_bound_port(lane, out); }
int coakka_v2_go_file_lane_prepare_receive(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_receive_spec_t *s) { return b->file_lane_prepare_receive(lane, s); }
int coakka_v2_go_file_lane_prepare_receive_grant(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_receive_spec_t *s, coakka_v2_file_receive_grant_t *out) { return b->file_lane_prepare_receive_grant(lane, s, out); }
int coakka_v2_go_file_lane_submit_send(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_send_spec_t *s) { return b->file_lane_submit_send(lane, s); }
int coakka_v2_go_file_lane_get_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d, coakka_v2_file_transfer_snapshot_t *out) { return b->file_lane_get_transfer(lane, id, d, out); }
int coakka_v2_go_file_lane_wait_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d, uint64_t seq, uint32_t ms, coakka_v2_file_transfer_snapshot_t *out) { return b->file_lane_wait_transfer(lane, id, d, seq, ms, out); }
int coakka_v2_go_file_lane_cancel_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d) { return b->file_lane_cancel_transfer(lane, id, d); }
int coakka_v2_go_file_lane_forget_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d) { return b->file_lane_forget_transfer(lane, id, d); }
int coakka_v2_go_file_lane_get_stats(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, coakka_v2_file_lane_stats_t *out) { return b->file_lane_get_stats(lane, out); }
int coakka_v2_go_file_sha256_path(coakka_v2_go_bindings_t *b, const char *path, uint8_t out[32], uint64_t *size) { return b->file_sha256_path(path, out, size); }

int coakka_v2_go_stream_lane_create_ex(coakka_v2_go_bindings_t *b, const coakka_v2_stream_lane_config_t *c, coakka_v2_stream_lane_t **out) { return b->stream_lane_create_ex(c, out); }
int coakka_v2_go_stream_lane_create_owned_ex(coakka_v2_go_bindings_t *b, const coakka_v2_stream_lane_owned_config_t *c, coakka_v2_stream_lane_t **out) { return b->stream_lane_create_owned_ex(c, out); }
void coakka_v2_go_stream_lane_destroy(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane) { b->stream_lane_destroy(lane); }
int coakka_v2_go_stream_lane_start(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane) { return b->stream_lane_start(lane); }
int coakka_v2_go_stream_lane_stop(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane) { return b->stream_lane_stop(lane); }
int coakka_v2_go_stream_lane_get_bound_port(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, uint16_t *out) { return b->stream_lane_get_bound_port(lane, out); }
int coakka_v2_go_stream_lane_prepare_publish(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, uint64_t format, uint32_t max_frame, uintptr_t context) {
  coakka_v2_stream_publish_spec_t spec = {sizeof(spec), id, token, format, max_frame, coakka_v2_go_stream_source_trampoline, (void *)context};
  return b->stream_lane_prepare_publish(lane, &spec);
}
int coakka_v2_go_stream_lane_prepare_publish_grant(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, uint64_t format, uint32_t max_frame, uintptr_t context, coakka_v2_stream_publish_grant_t *out) {
  coakka_v2_stream_publish_spec_t spec = {sizeof(spec), id, token, format, max_frame, coakka_v2_go_stream_source_trampoline, (void *)context};
  return b->stream_lane_prepare_publish_grant(lane, &spec, out);
}
int coakka_v2_go_stream_lane_subscribe(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, const char *host, uint16_t port, uint64_t format, uint32_t max_frame, uint32_t window, uint32_t timeout, uintptr_t context) {
  coakka_v2_stream_subscribe_spec_t spec = {sizeof(spec), id, token, host, port, format, max_frame, window, timeout, coakka_v2_go_stream_consumer_trampoline, (void *)context};
  return b->stream_lane_subscribe(lane, &spec);
}
int coakka_v2_go_stream_lane_get_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, coakka_v2_stream_session_snapshot_t *out) { return b->stream_lane_get_session(lane, id, direction, out); }
int coakka_v2_go_stream_lane_wait_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, uint64_t sequence, uint32_t timeout, coakka_v2_stream_session_snapshot_t *out) { return b->stream_lane_wait_session(lane, id, direction, sequence, timeout, out); }
int coakka_v2_go_stream_lane_get_pressure(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, coakka_v2_stream_pressure_snapshot_t *out) { return b->stream_lane_get_pressure(lane, id, direction, out); }
int coakka_v2_go_stream_lane_wait_pressure(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, uint64_t sequence, uint32_t timeout, coakka_v2_stream_pressure_snapshot_t *out) { return b->stream_lane_wait_pressure(lane, id, direction, sequence, timeout, out); }
int coakka_v2_go_stream_lane_cancel_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction) { return b->stream_lane_cancel_session(lane, id, direction); }
int coakka_v2_go_stream_lane_forget_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction) { return b->stream_lane_forget_session(lane, id, direction); }
int coakka_v2_go_stream_lane_get_stats(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, coakka_v2_stream_lane_stats_t *out) { return b->stream_lane_get_stats(lane, out); }
