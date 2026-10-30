#ifndef COAKKA_GO_RUNTIME_BRIDGE_H
#define COAKKA_GO_RUNTIME_BRIDGE_H

/* Private cgo boundary. Core owns every native record and callback type.
 * The module/table is opaque; Go owns binding lifetime and drains all native
 * calls/callbacks before close. Open is serialized by Go's nativeLibraryMu;
 * the process module/path is never retargeted or unloaded. Core byte owners never escape a copy/release
 * turn; packed staging is borrowed only for the synchronous cgo call. */
#include "coakka/v2/runtime.h"
#include "coakka/v2/runtime_startup.h"
#include "coakka/v2/client.h"
#include "coakka/v2/message_value.h"
#include "coakka/v2/control.h"
#include "coakka/v2/transport.h"
#include "coakka/v2/utils.h"
#include "coakka/v2/file_lane.h"
#include "coakka/v2/stream_lane.h"

typedef struct coakka_v2_go_bindings_t coakka_v2_go_bindings_t;

/* Private Windows loader representation bounds, including the wide NUL.
 * Not runtime policy or a promise of OS/filesystem long-path acceptance. */
enum {
  COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_UNITS = 32767,
  COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_BYTES =
      3 * (COAKKA_GO_WINDOWS_LIBRARY_PATH_MAX_UNITS - 1)
};

/* Language-owned packed spans, not a Core ABI record. A non-null large_packed
 * belongs to Go's boundary caller and is freed there after the final copy. */
typedef struct {
  size_t frame_len;
  size_t target_len;
  size_t payload_len;
  size_t packed_len;
  uint32_t one_way;
  uint8_t *large_packed;
} coakka_v2_go_packed_request_t;

/* By-value receipts avoid escaping Go output cells. These are private cgo
 * representations, not Core records or runtime limits. Inline bytes belong to
 * the returned value; a non-null owner is a Core byte lease that the Go caller
 * copies and releases. inline_owned describes storage only, never Core state. */
enum { COAKKA_GO_INLINE_PAYLOAD_BYTES = 256 };
typedef struct {
  size_t size;
  uint8_t *owner;
  uint32_t inline_owned;
  uint8_t inline_bytes[COAKKA_GO_INLINE_PAYLOAD_BYTES];
} coakka_v2_go_payload_value_t;

typedef struct {
  int status;
  uint32_t kind;
  coakka_v2_go_payload_value_t value;
} coakka_v2_go_payload_result_t;

typedef struct {
  int status;
  coakka_v2_go_packed_request_t value;
} coakka_v2_go_packed_result_t;

coakka_v2_go_bindings_t *coakka_v2_go_open_library(const char *path, char **error_out);
void coakka_v2_go_close_library(coakka_v2_go_bindings_t *bindings);
uint32_t coakka_v2_go_get_abi_version(coakka_v2_go_bindings_t *bindings);
int coakka_v2_go_get_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_info_t *out_info);
int coakka_v2_go_get_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_config_view_t *out_config);
int coakka_v2_go_get_capabilities(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_capability_snapshot_t *out_capabilities);
int coakka_v2_go_create_runtime_native(coakka_v2_go_bindings_t *bindings,
                                               const coakka_v2_runtime_config_t *cfg,
                                               const coakka_v2_tcp_connection_options_t *connection,
                                               const coakka_v2_startup_monitor_options_t *monitor,
                                               coakka_v2_runtime_t **out_runtime);
int coakka_v2_go_get_startup_config(coakka_v2_go_bindings_t *bindings,
                                            coakka_v2_runtime_t *runtime,
                                            coakka_v2_runtime_startup_config_t *out_config);
void coakka_v2_go_destroy_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt);
int coakka_v2_go_apply_network_options(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_network_options_t *options);
int coakka_v2_go_get_network_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_network_config_snapshot_t *out);
int coakka_v2_go_encode_message_value(coakka_v2_go_bindings_t *bindings, const coakka_v2_message_value_t *value, uint8_t **out, size_t *size);
int coakka_v2_go_decode_message_value(coakka_v2_go_bindings_t *bindings, const uint8_t *data, size_t size, coakka_v2_message_value_owner_t **owner, coakka_v2_message_value_t *out);
int coakka_v2_go_decode_deadletter_value(coakka_v2_go_bindings_t *bindings, const uint8_t *data, size_t size, coakka_v2_message_value_owner_t **owner, coakka_v2_deadletter_value_t *out);
void coakka_v2_go_message_value_release(coakka_v2_go_bindings_t *bindings, coakka_v2_message_value_owner_t *owner);
int coakka_v2_go_encode_control_value(coakka_v2_go_bindings_t *bindings, const coakka_v2_control_value_t *value, uint8_t **out, size_t *size);
int coakka_v2_go_get_host_handles(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_host_handles_t *out_handles);
int coakka_v2_go_start_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt);
int coakka_v2_go_stop_runtime(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt);
int coakka_v2_go_get_health(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_health_t *out_health);
int coakka_v2_go_get_stats(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_runtime_stats_t *out_stats);
int coakka_v2_go_submit_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len);
int coakka_v2_go_apply_control_envelope(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const uint8_t *buf, size_t len);
int coakka_v2_go_monitor_wait_consume(coakka_v2_go_bindings_t *bindings,
                                              int fd, uint32_t timeout_ms,
                                              uint64_t *out_signal_count);
void coakka_v2_go_host_handles_close(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_host_handles_t *handles);
int coakka_v2_go_request_reader_create(coakka_v2_go_bindings_t *bindings,
                                               coakka_v2_runtime_t *runtime,
                                               const coakka_v2_host_handles_t *handles,
                                               coakka_v2_request_reader_t **out_reader);
void coakka_v2_go_request_reader_destroy(coakka_v2_go_bindings_t *bindings,
                                                 coakka_v2_request_reader_t *reader);
int coakka_v2_go_ask_client_create_ex(
    coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *runtime,
    const coakka_v2_host_handles_t *handles, coakka_v2_ask_client_t **out_client);
void coakka_v2_go_ask_client_destroy(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_ask_client_t *client);
int coakka_v2_go_ask_client_begin(coakka_v2_go_bindings_t *bindings,
                                          coakka_v2_ask_client_t *client,
                                          const uint8_t *frame, size_t frame_len,
                                          coakka_v2_ask_ticket_t **out_ticket);
int coakka_v2_go_ask_ticket_await(coakka_v2_go_bindings_t *bindings,
                                          coakka_v2_ask_ticket_t *ticket,
                                          uint32_t timeout_ms, uint32_t *out_kind,
                                          uint8_t **out_frame, size_t *out_frame_len);
void coakka_v2_go_ask_ticket_cancel(coakka_v2_go_bindings_t *bindings,
                                            coakka_v2_ask_ticket_t *ticket);
void coakka_v2_go_ask_ticket_destroy(coakka_v2_go_bindings_t *bindings,
                                             coakka_v2_ask_ticket_t *ticket);
coakka_v2_go_packed_result_t coakka_v2_go_read_delivered_request_packed(
    coakka_v2_go_bindings_t *bindings, coakka_v2_request_reader_t *reader,
    uint32_t timeout_ms, uint8_t *staging, size_t staging_capacity);
coakka_v2_go_payload_result_t coakka_v2_go_ask_payload(
    coakka_v2_go_bindings_t *bindings, coakka_v2_ask_client_t *client,
    const char *message_id, const char *source, const char *target,
    const char *reply_to, const uint8_t *payload, size_t payload_len,
    uint32_t request_timeout_ms, uint32_t wait_timeout_ms,
    uint32_t delivery_hint);
int coakka_v2_go_submit_raw_reply(
    coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *runtime,
    const uint8_t *request_frame, size_t request_frame_len,
    const char *source, const uint8_t *payload, size_t payload_len);
void coakka_v2_go_client_bytes_release(coakka_v2_go_bindings_t *bindings,
                                               uint8_t *bytes);
int coakka_v2_go_apply_tcp_connection_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_connection_options_t *options, coakka_v2_tcp_connection_apply_result_t *out_result);
int coakka_v2_go_get_tcp_connection_config(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_connection_config_snapshot_t *out_config);
int coakka_v2_go_apply_tcp_security_options_ex(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, const coakka_v2_tcp_security_options_t *options, coakka_v2_tcp_security_apply_result_t *out_result);
int coakka_v2_go_get_tcp_security_info(coakka_v2_go_bindings_t *bindings, coakka_v2_runtime_t *rt, coakka_v2_tcp_security_info_t *out_info);
const char *coakka_v2_go_transport_apply_reason_name(coakka_v2_go_bindings_t *bindings, uint32_t reason);
int coakka_v2_go_query_feature(coakka_v2_go_bindings_t *bindings, uint32_t feature, uint32_t *out_supported);
int coakka_v2_go_file_lane_create_ex(coakka_v2_go_bindings_t *b, const coakka_v2_file_lane_config_t *c, coakka_v2_file_lane_t **out);
int coakka_v2_go_file_lane_create_owned_ex(coakka_v2_go_bindings_t *b, const coakka_v2_file_lane_owned_config_t *c, coakka_v2_file_lane_t **out);
void coakka_v2_go_file_lane_destroy(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane);
int coakka_v2_go_file_lane_start(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane);
int coakka_v2_go_file_lane_stop(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane);
int coakka_v2_go_file_lane_get_bound_port(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, uint16_t *out);
int coakka_v2_go_file_lane_prepare_receive(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_receive_spec_t *s);
int coakka_v2_go_file_lane_prepare_receive_grant(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_receive_spec_t *s, coakka_v2_file_receive_grant_t *out);
int coakka_v2_go_file_lane_submit_send(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const coakka_v2_file_send_spec_t *s);
int coakka_v2_go_file_lane_get_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d, coakka_v2_file_transfer_snapshot_t *out);
int coakka_v2_go_file_lane_wait_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d, uint64_t seq, uint32_t ms, coakka_v2_file_transfer_snapshot_t *out);
int coakka_v2_go_file_lane_cancel_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d);
int coakka_v2_go_file_lane_forget_transfer(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, const char *id, uint32_t d);
int coakka_v2_go_file_lane_get_stats(coakka_v2_go_bindings_t *b, coakka_v2_file_lane_t *lane, coakka_v2_file_lane_stats_t *out);
int coakka_v2_go_file_sha256_path(coakka_v2_go_bindings_t *b, const char *path, uint8_t out[32], uint64_t *size);
int coakka_v2_go_stream_lane_create_ex(coakka_v2_go_bindings_t *b, const coakka_v2_stream_lane_config_t *c, coakka_v2_stream_lane_t **out);
int coakka_v2_go_stream_lane_create_owned_ex(coakka_v2_go_bindings_t *b, const coakka_v2_stream_lane_owned_config_t *c, coakka_v2_stream_lane_t **out);
void coakka_v2_go_stream_lane_destroy(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane);
int coakka_v2_go_stream_lane_start(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane);
int coakka_v2_go_stream_lane_stop(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane);
int coakka_v2_go_stream_lane_get_bound_port(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, uint16_t *out);
int coakka_v2_go_stream_lane_prepare_publish(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, uint64_t format, uint32_t max_frame, uintptr_t context);
int coakka_v2_go_stream_lane_prepare_publish_grant(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, uint64_t format, uint32_t max_frame, uintptr_t context, coakka_v2_stream_publish_grant_t *out);
int coakka_v2_go_stream_lane_subscribe(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, const char *token, const char *host, uint16_t port, uint64_t format, uint32_t max_frame, uint32_t window, uint32_t timeout, uintptr_t context);
int coakka_v2_go_stream_lane_get_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, coakka_v2_stream_session_snapshot_t *out);
int coakka_v2_go_stream_lane_wait_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, uint64_t sequence, uint32_t timeout, coakka_v2_stream_session_snapshot_t *out);
int coakka_v2_go_stream_lane_get_pressure(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, coakka_v2_stream_pressure_snapshot_t *out);
int coakka_v2_go_stream_lane_wait_pressure(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction, uint64_t sequence, uint32_t timeout, coakka_v2_stream_pressure_snapshot_t *out);
int coakka_v2_go_stream_lane_cancel_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction);
int coakka_v2_go_stream_lane_forget_session(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, const char *id, uint32_t direction);
int coakka_v2_go_stream_lane_get_stats(coakka_v2_go_bindings_t *b, coakka_v2_stream_lane_t *lane, coakka_v2_stream_lane_stats_t *out);

#endif
