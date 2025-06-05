package executor_test

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHCmdExecutor", func() {
	var (
		ctx       context.Context
		exec      *executor.SSHCmdExecutor
		task      model.Task
		errAction error
	)

	BeforeEach(func() {
		ctx = context.Background()
		exec = executor.NewSSHCmdExecutor()
	})

	JustBeforeEach(func() {
		errAction = exec.Run(ctx, task)
	})

	// TODO: Add happy path test coverage

	When("the command parameter is missing", func() {
		BeforeEach(func() {
			task = model.Task{
				Name:       "missing-command",
				Parameters: map[string]string{},
			}
		})

		It("returns an error", func() {
			Expect(errAction.Error()).To(ContainSubstring("missing 'command' parameter"))
		})
	})

	When("the SSH config is invalid", func() {
		BeforeEach(func() {
			task = model.Task{
				Name: "invalid-ssh-config",
				Parameters: map[string]string{
					"command": "uptime",
					"user":    "user",
					"host":    "localhost",
					"port":    "22",
					// password or keyPath missing
				},
			}
		})

		It("returns a config error", func() {
			Expect(errAction.Error()).To(ContainSubstring("either password or keyPath must be provided"))
		})
	})

	When("SSH connection fails", func() {
		BeforeEach(func() {
			task = model.Task{
				Name: "bad-ssh-connection",
				Parameters: map[string]string{
					"command":  "ls -la",
					"user":     "user",
					"host":     "localhost",
					"port":     "2222", // assumed to be unreachable
					"password": "wrong",
				},
			}
		})

		It("returns a connection error", func() {
			Expect(errAction.Error()).To(ContainSubstring("failed to dial SSH"))
		})
	})
})
