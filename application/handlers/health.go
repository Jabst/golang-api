package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
)

type runtimeCheckResponse struct {
	TotalAllocatedMemory uint64 `json:"total_allocated_memory_MB"`
	AllocatedMemory      uint64 `json:"allocated_memory_MB"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {

	_, err := w.Write([]byte("OK"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}
	w.WriteHeader(http.StatusOK)
}

func RuntimeCheck(w http.ResponseWriter, r *http.Request) {

	var res runtimeCheckResponse

	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)

	res.TotalAllocatedMemory = (memstats.TotalAlloc / 1024 / 1024)
	res.AllocatedMemory = (memstats.Alloc / 1024 / 1024)

	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
}
