package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/handler"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/handler/handlerfakes"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Handlers", func() {
	var (
		e      *echo.Echo
		ctx    context.Context
		rec    *httptest.ResponseRecorder
		fakePS *handlerfakes.FakeProcessStore
		id     uuid.UUID
		req    *http.Request
	)

	BeforeEach(func() {
		e = echo.New()
		ctx = context.Background()
		rec = httptest.NewRecorder()
		fakePS = &handlerfakes.FakeProcessStore{}
		handler.RegisterHandlers(ctx, e, fakePS)
		id = uuid.New()
	})

	JustBeforeEach(func() {
		e.ServeHTTP(rec, req)
	})

	Describe("GET /listProcesses", func() {
		BeforeEach(func() {
			expected := []model.ProcessRun{
				{ID: id, Definition: model.ProcessDefinition{Name: "test"}, Status: model.StatusRunning, StartedAt: time.Now()},
			}
			fakePS.ListRunningProcessesReturns(expected, nil)

			req = httptest.NewRequest(http.MethodGet, "/listProcesses", nil)
		})

		It("returns list of running processes", func() {
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result []model.ProcessRun
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(Succeed())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ID).To(Equal(id))
		})
	})

	Describe("GET /listProcess/:id", func() {
		BeforeEach(func() {
			expected := model.ProcessRun{ID: id, Definition: model.ProcessDefinition{Name: "test"}, Status: model.StatusRunning}
			fakePS.GetProcessByIDReturns(expected, nil)

			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/listProcess/%s", id), nil)
		})

		It("returns the process by ID", func() {
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result model.ProcessRun
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(Succeed())
			Expect(result.ID).To(Equal(id))
		})
	})

	Describe("POST /stopProcess/:id", func() {
		BeforeEach(func() {
			fakePS.UpdateProcessStatusReturns(nil)

			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/stopProcess/%s", id), nil)
		})

		It("updates process status to stopped", func() {
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(fakePS.UpdateProcessStatusCallCount()).To(Equal(1))
			_, calledID, status := fakePS.UpdateProcessStatusArgsForCall(0)
			Expect(calledID).To(Equal(id))
			Expect(status).To(Equal(model.StatusStopped))
		})
	})

	Describe("GET /processlog/:id", func() {
		BeforeEach(func() {
			logs := []model.ProcessLog{
				{ID: 1, ProcessID: id, Log: "started"},
			}
			fakePS.GetProcessLogsReturns(logs, nil)

			req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/processlog/%s", id), nil)

		})

		It("returns process logs", func() {
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result []model.ProcessLog
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(Succeed())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Log).To(Equal("started"))
		})
	})
})
