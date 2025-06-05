package executor_test

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SCPCmdExecutor", func() {
	var (
		ctx       context.Context
		exec      *executor.SCPCmdExecutor
		task      model.Task
		errAction error
	)

	BeforeEach(func() {
		ctx = context.Background()
		exec = executor.NewSCPCmdExecutor()
	})

	JustBeforeEach(func() {
		errAction = exec.Run(ctx, task)
	})

	// TODO: Add happy path test coverage.

	When("missing parameters", func() {
		BeforeEach(func() {
			task = model.Task{
				Name:       "missing-paths",
				Parameters: map[string]string{},
			}
		})
		It("should return error if localPath or remotePath are missing", func() {
			Expect(errAction.Error()).To(ContainSubstring("missing 'localPath' or 'remotePath'"))
		})
	})

	When("sshutil returns error", func() {
		BeforeEach(func() {
			task = model.Task{
				Name: "bad-ssh-config",
				Parameters: map[string]string{
					"localPath":  "somefile",
					"remotePath": "/tmp/target",
					"user":       "u",
					"host":       "h",
					"port":       "22",
					// missing password/keyPath
				},
			}
		})

		It("should fail if config building fails", func() {
			Expect(errAction.Error()).To(ContainSubstring("either password or keyPath must be provided"))
		})
	})
})
