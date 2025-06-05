package executor_test

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalCmdExecutor", func() {
	var (
		executorSvc *executor.LocalCmdExecutor
		ctx         context.Context
		task        model.Task
		errAction   error
	)

	JustBeforeEach(func() {
		errAction = executorSvc.Run(ctx, task)
	})

	BeforeEach(func() {
		executorSvc = executor.NewLocalCmdService()
		ctx = context.Background()

		task = model.Task{
			Name: "echo test",
			Parameters: map[string]string{
				"command": "echo 123",
			},
		}
	})

	It("suceeeds", func() {
		Expect(errAction).ToNot(HaveOccurred())
	})

	Context("with a missing command parameter", func() {
		BeforeEach(func() {
			task = model.Task{
				Name:       "missing-command",
				Parameters: map[string]string{},
			}

		})
		It("should return an error", func() {
			Expect(errAction.Error()).To(ContainSubstring("missing 'command' parameter"))
		})
	})

	Context("with an invalid shell command", func() {
		BeforeEach(func() {
			task = model.Task{
				Name: "invalid-cmd",
				Parameters: map[string]string{
					"command": "nonexistentcommand",
				},
			}
		})

		It("should return an error", func() {
			Expect(errAction.Error()).To(ContainSubstring("local command execution failed"))
		})
	})
})
