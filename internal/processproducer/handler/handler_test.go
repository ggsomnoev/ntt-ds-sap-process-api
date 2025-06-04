package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/handler"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/handler/handlerfakes"
)

var ErrPublishFailed = errors.New("failed to publish process")

var _ = Describe("Process Handler", func() {
	var (
		e         *echo.Echo
		ctx       context.Context
		recorder  *httptest.ResponseRecorder
		publisher *handlerfakes.FakePublisher
	)

	BeforeEach(func() {
		e = echo.New()
		ctx = context.Background()
		recorder = httptest.NewRecorder()
		publisher = &handlerfakes.FakePublisher{}
		handler.RegisterHandlers(ctx, e, publisher)
	})

	Describe("POST /startProcess", func() {
		var (
			process model.ProcessDefinition
			req     *http.Request
		)

		BeforeEach(func() {
			process = model.ProcessDefinition{
				Name: "sample",
				Params: []model.Param{
					{Name: "env", Mandatory: true, DefValue: "dev", Description: "Environment"},
				},
				Tasks: []model.Task{
					{Name: "print", Class: model.LocalCmd, Parameters: map[string]string{"cmd": "echo Hello"}},
				},
			}

			body, _ := json.Marshal(process)
			req = httptest.NewRequest(http.MethodPost, "/startProcess", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		})

		JustBeforeEach(func() {
			e.ServeHTTP(recorder, req)
		})

		It("succeeds", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(publisher.PublishCallCount()).To(Equal(1))
			_, actualMessage := publisher.PublishArgsForCall(0)
			Expect(actualMessage.UUID).NotTo(Equal(uuid.Nil))
			Expect(actualMessage.ProcessDefinition).To(Equal(process))
		})

		When("invalid JSON is posted", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodPost, "/startProcess", bytes.NewBufferString("{invalid json"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			})

			It("returns 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		When("validation fails", func() {
			BeforeEach(func() {
				process = model.ProcessDefinition{
					Name:  "",
					Tasks: []model.Task{},
				}
				body, _ := json.Marshal(process)
				req = httptest.NewRequest(http.MethodPost, "/startProcess", bytes.NewReader(body))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			})

			It("returns 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		When("publish fails", func() {
			BeforeEach(func() {
				publisher.PublishReturns(ErrPublishFailed)
			})

			It("returns 500", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})
