package store_test

import (
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/store"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessDBStore", func() {
	var (
		s         *store.ProcessDBStore
		runID     uuid.UUID
		run       model.ProcessRun
		errAction error
	)

	BeforeEach(func() {
		s = store.NewProcessDBStore(pool)

		// Clean tables before test. TODO: adjust to proper cleanup in after each / just after each.
		_, _ = pool.Exec(ctx, "DELETE FROM process_logs")
		_, _ = pool.Exec(ctx, "DELETE FROM process_runs")

		runID = uuid.New()

		run = model.ProcessRun{
			ID:         runID,
			Definition: model.ProcessDefinition{Name: "my-process"},
			Status:     model.StatusRunning,
			StartedAt:  time.Now(),
		}
	})

	Describe("InsertProcess", func() {
		JustBeforeEach(func() {
			errAction = s.InsertProcess(ctx, run)
		})

		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})
	})

	Describe("UpdateProcessStatus", func() {
		BeforeEach(func() {
			Expect(s.InsertProcess(ctx, run)).To(Succeed())
		})

		JustBeforeEach(func() {
			errAction = s.UpdateProcessStatus(ctx, runID, model.StatusCompleted)
		})

		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})

		It("sets proper status", func() {
			stored, err := s.GetProcessByID(ctx, runID)
			Expect(err).ToNot(HaveOccurred())
			Expect(stored.Status).To(Equal(model.StatusCompleted))
			Expect(stored.EndedAt).ToNot(BeNil())
		})
	})

	Describe("GetProcessByID", func() {
		var result model.ProcessRun

		BeforeEach(func() {
			Expect(s.InsertProcess(ctx, run)).To(Succeed())
		})

		JustBeforeEach(func() {
			result, errAction = s.GetProcessByID(ctx, runID)
		})

		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})

		It("retrieves the inserted process", func() {
			Expect(result.ID).To(Equal(runID))
			Expect(result.Definition).To(Equal(run.Definition))
		})
	})

	Describe("ListRunningProcesses", func() {
		var results []model.ProcessRun

		BeforeEach(func() {
			for i := 0; i < 3; i++ {
				run := model.ProcessRun{
					ID:         uuid.New(),
					Definition: model.ProcessDefinition{Name: "my-process"},
					Status:     model.StatusRunning,
					StartedAt:  time.Now().Add(time.Duration(-i) * time.Minute),
				}
				Expect(s.InsertProcess(ctx, run)).To(Succeed())
			}
		})

		JustBeforeEach(func() {
			results, errAction = s.ListRunningProcesses(ctx)
		})

		It("succeeds", func() {
			Expect(errAction).ToNot(HaveOccurred())
		})

		It("returns all running processes", func() {
			Expect(results).To(HaveLen(3))
			Expect(results[0].StartedAt.After(results[1].StartedAt)).To(BeTrue()) // Sorted DESC
		})
	})

	Describe("AppendProcessLog and GetProcessLogs", func() {
		BeforeEach(func() {
			Expect(s.InsertProcess(ctx, run)).To(Succeed())
		})

		JustBeforeEach(func() {
			Expect(s.AppendProcessLog(ctx, runID, "log 1")).To(Succeed())
			Expect(s.AppendProcessLog(ctx, runID, "log 2")).To(Succeed())
		})

		It("fetches logs", func() {
			logs, err := s.GetProcessLogs(ctx, runID)
			Expect(err).ToNot(HaveOccurred())
			Expect(logs).To(HaveLen(2))
			Expect(logs[0].Log).To(Equal("log 1"))
			Expect(logs[1].Log).To(Equal("log 2"))
		})
	})
})
