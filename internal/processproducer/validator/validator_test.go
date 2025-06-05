package validator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/validator"
)

var _ = Describe("ProcessValidator", func() {
	var (
		validatorSvc *validator.ProcessValidator
		proc         model.ProcessDefinition
		err          error
	)

	BeforeEach(func() {
		validatorSvc = validator.NewProcessValidator()
	})

	Describe("Validate", func() {
		BeforeEach(func() {
			proc = model.ProcessDefinition{
				Name: "valid-process",
				Params: []model.Param{
					{Name: "param1", Mandatory: true},
				},
				Tasks: []model.Task{
					{Name: "task1", Class: "some.class", WaitFor: nil},
					{Name: "task2", Class: "other.class", WaitFor: []string{"task1"}},
				},
			}
		})

		JustBeforeEach(func() {
			err = validatorSvc.Validate(proc)
		})

		It("succeeds", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		When("process name is empty", func() {
			BeforeEach(func() {
				proc.Name = " "
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("process name must not be empty"))
			})
		})

		When("a task name is empty", func() {
			BeforeEach(func() {
				proc.Tasks[0].Name = " "
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("task name must not be empty"))
			})
		})

		When("a task class is empty", func() {
			BeforeEach(func() {
				proc.Tasks[0].Class = " "
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("task 'task1' class must not be empty"))
			})
		})

		When("duplicate task names exist", func() {
			BeforeEach(func() {
				proc.Tasks = append(proc.Tasks, model.Task{Name: "task1", Class: "dup.class"})
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("duplicate task name found: task1"))
			})
		})

		When("a task waits for itself", func() {
			BeforeEach(func() {
				proc.Tasks[0].WaitFor = []string{"task1"}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("task 'task1' cannot wait for itself"))
			})
		})

		When("a param name is empty", func() {
			BeforeEach(func() {
				proc.Params[0].Name = " "
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("param name must not be empty"))
			})
		})

		When("duplicate param names exist", func() {
			BeforeEach(func() {
				proc.Params = append(proc.Params, model.Param{Name: "param1"})
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("duplicate param name found: param1"))
			})
		})

		When("a task waits for an unknown task", func() {
			BeforeEach(func() {
				proc.Tasks[1].WaitFor = []string{"nonexistent"}
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("task 'task2' waits for unknown task 'nonexistent'"))
			})
		})
	})

	Describe("ValidateMandatoryParams", func() {
		var params map[string]string

		BeforeEach(func() {
			proc = model.ProcessDefinition{
				Name: "ssh-process",
				Params: []model.Param{
					{Name: "host", Mandatory: true},
					{Name: "port", Mandatory: true},
					{Name: "user", Mandatory: true},
					{Name: "password", Mandatory: true},
					{Name: "optional1", Mandatory: false},
				},
			}
		})

		JustBeforeEach(func() {
			err = validatorSvc.ValidateMandatoryParams(proc, params)
		})

		When("all mandatory parameters are provided", func() {
			BeforeEach(func() {
				params = map[string]string{
					"host":     "example.com",
					"port":     "22",
					"user":     "admin",
					"password": "secret",
				}
			})

			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("one mandatory parameter is missing", func() {
			BeforeEach(func() {
				params = map[string]string{
					"host": "example.com",
					"port": "22",
					"user": "admin",
					// password is missing
				}
			})

			It("returns an error mentioning the missing parameter", func() {
				Expect(err).To(MatchError("missing mandatory parameters: password"))
			})
		})

		When("multiple mandatory parameters are missing", func() {
			BeforeEach(func() {
				params = map[string]string{
					"host": "example.com",
					// port, user, password are missing
				}
			})

			It("returns an error listing all missing params", func() {
				Expect(err).To(MatchError(ContainSubstring("missing mandatory parameters:")))
				Expect(err.Error()).To(ContainSubstring("port"))
				Expect(err.Error()).To(ContainSubstring("user"))
				Expect(err.Error()).To(ContainSubstring("password"))
			})
		})

		When("a mandatory parameter is present but blank", func() {
			BeforeEach(func() {
				params = map[string]string{
					"host":     "example.com",
					"port":     "22",
					"user":     "   ", // blank
					"password": "secret",
				}
			})

			It("returns an error about the blank parameter", func() {
				Expect(err).To(MatchError("missing mandatory parameters: user"))
			})
		})

		When("optional parameter is missing", func() {
			BeforeEach(func() {
				params = map[string]string{
					"host":     "example.com",
					"port":     "22",
					"user":     "admin",
					"password": "secret",
					// optional1 missing
				}
			})

			It("still succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
