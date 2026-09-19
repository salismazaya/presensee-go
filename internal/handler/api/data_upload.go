package api

import (
	"encoding/json"
	"net/http"

	"presensee/internal/middleware"
	"presensee/internal/service/dump"
	"presensee/internal/service/lzstring"
	"presensee/internal/service/sync"

	"gorm.io/gorm"
)

type SyncHandler struct {
	DB *gorm.DB
}

func (h *SyncHandler) GetData(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sqlDump, err := dump.GenerateDump(h.DB, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(sqlDump))
}

type CompressedUploadRequest struct {
	Data string `json:"data"`
}

func (h *SyncHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	var payload sync.UploadPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Payload tidak valid"})
		return
	}

	res, err := sync.ProcessSync(h.DB, user, payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": res})
}

func (h *SyncHandler) CompressedUpload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Unauthorized"})
		return
	}

	var req CompressedUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Format request tidak valid"})
		return
	}

	decompressed, err := lzstring.DecompressFromBase64(req.Data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Gagal dekompresi data"})
		return
	}

	var actionItems []sync.ActionItem
	if err := json.Unmarshal([]byte(decompressed), &actionItems); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Gagal parsing isi data terkompresi"})
		return
	}

	payload := sync.UploadPayload{Data: actionItems}
	res, err := sync.ProcessSync(h.DB, user, payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": res})
}
