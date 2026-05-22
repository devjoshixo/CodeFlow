package execution

import (
	"encoding/json"
	"net/http"
	"time"
)

type ExecutionHandler struct {
	execution *ExecutionService
}

type ExecutionResponse struct {
	ID         string    `json:"id"`
	Language   string    `json:"language"`
	Code       string    `json:"code"`
	Status     string    `json:"status"`
	Output     string    `json:"output"`
	Error      string    `json:"error"`
	ExitCode   int       `json:"exit_code"`
	DurationMs int       `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewExecutionHandler(execution *ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{execution: execution}
}

func decode[T any](r *http.Request, v *T) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var UserID = "4a6ba473-79bd-4fa9-b7bd-08fbe8f54263"

func (h *ExecutionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}

	if err := decode(r, &body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	exec := &Execution{
		UserID:   UserID, // extracted from JWT context
		Language: body.Language,
		Code:     body.Code,
	}
	submission, err := h.execution.Submit(r.Context(), exec)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respond(w, http.StatusCreated, map[string]any{"id": submission.ID, "status": submission.Status})
}

func (h *ExecutionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	exec, err := h.execution.GetByID(r.Context(), id, UserID)
	if err != nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "no submissions found with this id"})
		return
	}

	body := &ExecutionResponse{}

	body.ID = exec.ID
	body.Language = exec.Language
	body.Code = exec.Code
	body.Status = string(exec.Status)
	body.Output = strVal(exec.Output)
	body.Error = strVal(exec.Error)
	body.ExitCode = exec.ExitCode
	body.DurationMs = exec.DurationMs
	body.CreatedAt = exec.CreatedAt

	respond(w, http.StatusOK, map[string]any{"message": "submission found", "execution": (body)})

}

func (h *ExecutionHandler) GetByUser(w http.ResponseWriter, r *http.Request) {

	execs, err := h.execution.GetByUser(r.Context(), UserID)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var bodies []ExecutionResponse

	for _, exec := range execs {
		body := ExecutionResponse{
			ID:         exec.ID,
			Language:   exec.Language,
			Code:       exec.Code,
			Status:     string(exec.Status),
			Output:     strVal(exec.Output),
			Error:      strVal(exec.Error),
			ExitCode:   exec.ExitCode,
			DurationMs: exec.DurationMs,
			CreatedAt:  exec.CreatedAt,
		}

		bodies = append(bodies, body)
	}

	respond(w, http.StatusOK, map[string]any{
		"message":   "submission found",
		"execution": bodies,
	})

}
