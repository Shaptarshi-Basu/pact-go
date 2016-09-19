package native

/*
#cgo LDFLAGS: ${SRCDIR}/../../libs/libpact_mock_server.dylib

// Library headers
typedef int bool;
#define true 1
#define false 0

int create_mock_server(char* pact, int port);
int mock_server_matched(int port);
char* mock_server_mismatches(int port);
bool cleanup_mock_server(int port);
int write_pact_file(int port, char* dir);

*/
import "C"
import (
	"encoding/json"
	"log"
)

// Request is the sub-struct of Mismatch
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// Mismatch is a type returned from the validation process
//
// [
//   {
//     "method": "GET",
//     "path": "/",
//     "request": {
//       "body": {
//         "pass": 1234,
//         "user": {
//           "address": "some address",
//           "name": "someusername",
//           "phone": 12345678,
//           "plaintext": "plaintext"
//         }
//       },
//       "method": "GET",
//       "path": "/"
//     },
//     "type": "missing-request"
//   }
// ]
type Mismatch struct {
	Request Request
	Type    string
}

// CreateMockServer creates a new Mock Server from a given Pact file.
func CreateMockServer(pact string) int {
	log.Println("[DEBUG] mock server starting")
	res := C.create_mock_server(C.CString(pact), 0)
	log.Println("[DEBUG] mock server running on port:", res)
	return int(res)
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es.
func Verify(port int, dir string) (bool, []Mismatch) {
	res := C.mock_server_matched(C.int(port))
	defer CleanupMockServer(port)

	mismatches := MockServerMismatches(port)
	log.Println("[DEBUG] mock server mismatches:", len(mismatches))

	if int(res) == 1 {
		log.Println("[DEBUG] mock server write pact file")
		WritePactFile(port, dir)
	}

	return int(res) == 1, mismatches
}

// MockServerMismatches returns a JSON object containing any mismatches from
// the last set of interactions.
func MockServerMismatches(port int) []Mismatch {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []Mismatch

	mismatches := C.mock_server_mismatches(C.int(port))
	json.Unmarshal([]byte(C.GoString(mismatches)), &res)

	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func CleanupMockServer(port int) {
	log.Println("[DEBUG] mock server cleaning up port:", port)
	C.cleanup_mock_server(C.int(port))
}

// WritePactFile writes the Pact to file.
func WritePactFile(port int, dir string) int {
	log.Println("[DEBUG] pact verify on port:", port, ", dir:", dir)
	return int(C.write_pact_file(C.int(port), C.CString(dir)))
}
