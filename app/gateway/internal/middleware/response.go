package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"flamingo/pkg/errors"
	pkggrpc "flamingo/pkg/grpc"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// BusinessResponse is the unified API response envelope.
type BusinessResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// MarshalJSON serializes the response envelope to JSON. When Data is a protobuf
// message, it uses protojson so that well-known types (e.g. google.protobuf.Timestamp)
// are emitted as their canonical JSON form (RFC3339 strings for Timestamps) rather
// than the raw {"seconds":...,"nanos":...} struct shape that encoding/json produces.
func (r BusinessResponse) MarshalJSON() ([]byte, error) {
	// Write the leading fields into the encoder first, keeping the data field out
	// so we can control how it is rendered.
	shell := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{Code: r.Code, Message: r.Message}
	buf, err := json.Marshal(shell)
	if err != nil {
		return nil, err
	}

	var body []byte
	switch d := r.Data.(type) {
	case nil:
		body, err = json.Marshal(d)
	case proto.Message:
		body, err = protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(d)
	case json.Marshaler:
		body, err = d.MarshalJSON()
	default:
		body, err = json.Marshal(d)
	}
	if err != nil {
		return nil, err
	}

	// Assemble the full JSON object: { "code":..., "message":..., "data":... }.
	// buf is like: {"code":0,"message":"success"}
	// We need to insert the data field before the closing brace.
	var out bytes.Buffer
	out.Grow(len(buf) + len(body) + 8)
	// drop trailing brace
	head := buf[:len(buf)-1]
	out.Write(head)
	out.WriteString(`,"data":`)
	out.Write(body)
	out.WriteByte('}')
	return out.Bytes(), nil
}

// ResponseEncoder is the Kratos ResponseEncoder that wraps success responses
// in the {code, message, data} envelope.
func ResponseEncoder(w http.ResponseWriter, r *http.Request, v any) error {
	resp := BusinessResponse{
		Code:    0,
		Message: "success",
		Data:    v,
	}
	return writeJSON(w, http.StatusOK, resp)
}

// ErrorEncoder is the Kratos ErrorEncoder that wraps error responses
// in the {code, message, null} envelope.
func ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	bizCode, msg := extractErrorInfo(err)
	if bizCode == 0 {
		bizCode = errors.CodeInternalError
	}
	httpStatus := mapBusinessCodeToHTTP(bizCode, err)

	// X-Error-Code header is already set by ErrorCodeMiddleware; set it here
	// as a safety net for errors that bypass the middleware chain.
	w.Header().Set("X-Error-Code", strconv.Itoa(bizCode))

	resp := BusinessResponse{
		Code:    bizCode,
		Message: msg,
		Data:    nil,
	}
	_ = writeJSON(w, httpStatus, resp)
}

// extractErrorInfo extracts the business error code and message from a gRPC error.
func extractErrorInfo(err error) (int, string) {
	bizCode := pkggrpc.ExtractBusinessCode(err)

	var msg string
	if st, ok := status.FromError(err); ok {
		msg = st.Message()
	}
	if msg == "" {
		msg = err.Error()
	}

	if bizCode == 0 {
		return 0, msg
	}

	// Prefer the localized message from the error message registry.
	codeMsg := errors.GetMessage(bizCode)
	if codeMsg != "" {
		return bizCode, codeMsg
	}
	return bizCode, msg
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
