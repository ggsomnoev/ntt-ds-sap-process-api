package service_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/servicefakes"
)

var (
	ErrDb            = errors.New("db error")
	ErrInsert        = errors.New("insert error")
	ErrMarkCompleted = errors.New("mark completed error")
)

var _ = Describe("Service", func() {
	When("created", func() {
		It("should instantiate the service", func() {
			Expect(service.NewService(nil, nil, nil)).NotTo(BeNil())
		})
	})

	Describe("Run", func() {
		var (
			svc          *service.Service
			ctx          context.Context
			store        *servicefakes.FakeStore
			processStore *servicefakes.FakeProcessStore
			executor     *servicefakes.FakeExecutor
			executors    map[model.ClassType]service.Executor
			msg          model.Message
			errAction    error
		)

		BeforeEach(func() {
			ctx = context.Background()
			store = &servicefakes.FakeStore{}
			processStore = &servicefakes.FakeProcessStore{}
			executor = &servicefakes.FakeExecutor{}
			executors = map[model.ClassType]service.Executor{
				"someCmd": executor,
			}
			svc = service.NewService(store, processStore, executors)

			msg = model.Message{
				UUID: uuid.New(),
				ProcessDefinition: model.ProcessDefinition{
					Name: "TestProcess",
				},
			}

			store.RunInAtomicallyStub = func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}

			store.MessageExistsReturns(false, nil)
			store.AddMessageReturns(nil)
			store.MarkCompletedReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.Run(ctx, msg)
		})

		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())

			Expect(store.MessageExistsCallCount()).To(Equal(1))
			Expect(store.AddMessageCallCount()).To(Equal(1))
			Expect(store.MarkCompletedCallCount()).To(Equal(1))
		})

		When("the message exists", func() {
			BeforeEach(func() {
				store.MessageExistsReturns(true, nil)
			})

			It("should skip processing the message", func() {
				Expect(store.AddMessageCallCount()).To(Equal(0))
			})
		})

		When("MessageExists fails", func() {
			BeforeEach(func() {
				store.MessageExistsReturns(false, ErrDb)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrDb))
			})
		})

		When("AddMessage fails", func() {
			BeforeEach(func() {
				store.AddMessageReturns(ErrInsert)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrInsert))
			})
		})

		When("MarkCompleted fails", func() {
			BeforeEach(func() {
				store.MarkCompletedReturns(ErrMarkCompleted)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrMarkCompleted))
			})
		})
	})
})
