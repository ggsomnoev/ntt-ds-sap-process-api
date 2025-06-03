package store_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/store"
)

var _ = Describe("Store", func() {
	When("created", func() {
		It("exists", func() {
			Expect(store.NewStore(nil)).NotTo(BeNil())
		})
	})

	var _ = Describe("instance", Serial, func() {
		var (
			s         *store.Store
			errAction error
			filename  string
		)

		BeforeEach(func() {
			s = store.NewStore(pool)

			filename = "test.yaml"
		})

		Describe("AddProcessedFile", func() {
			JustBeforeEach(func() {
				errAction = s.RunInAtomically(ctx, func(ctx context.Context) error {
					return s.AddProcessedFile(ctx, filename)
				})
			})

			JustAfterEach(func() {
				err := s.DeleteProcessedFile(ctx, filename)
				Expect(err).NotTo(HaveOccurred())
			})

			It("succeeds", func() {
				Expect(errAction).NotTo(HaveOccurred())
			})

			Context("and the filename is inserted", func() {
				var (
					exists bool
					err    error
				)

				BeforeEach(func() {
					errAction = s.RunInAtomically(ctx, func(ctx context.Context) error {
						exists, err = s.FileExists(ctx, filename)
						return nil
					})
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(exists).To(BeTrue())
				})

			})
		})

		Describe("MarkCompleted", func() {
			BeforeEach(func() {
				err := s.RunInAtomically(ctx, func(ctx context.Context) error {
					return s.AddProcessedFile(ctx, filename)
				})
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := s.DeleteProcessedFile(ctx, filename)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				errAction = s.RunInAtomically(ctx, func(txCtx context.Context) error {
					return s.MarkCompleted(txCtx, filename)
				})
			})

			It("succeeds", func() {
				Expect(errAction).NotTo(HaveOccurred())
			})

			It("marks the message as completed", func() {
				completedAt, err := s.GetCompletedAtByFilename(ctx, filename)
				Expect(err).NotTo(HaveOccurred())
				Expect(completedAt).NotTo(BeZero())
			})
		})

		Describe("FileExists", func() {
			var exists bool
			JustBeforeEach(func() {
				errAction = s.RunInAtomically(ctx, func(txCtx context.Context) error {
					var err error
					exists, err = s.FileExists(txCtx, filename)
					return err
				})
			})

			It("returns false for filename", func() {
				Expect(exists).To(BeFalse())
			})

			Context("and a filename is added", func() {
				var exists bool
				BeforeEach(func() {
					err := s.RunInAtomically(ctx, func(txCtx context.Context) error {
						return s.AddProcessedFile(txCtx, filename)
					})
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err := s.DeleteProcessedFile(ctx, filename)
					Expect(err).NotTo(HaveOccurred())
				})

				JustBeforeEach(func() {
					errAction = s.RunInAtomically(ctx, func(txCtx context.Context) error {
						var err error
						exists, err = s.FileExists(txCtx, filename)
						return err
					})
				})

				It("succeeds", func() {
					Expect(errAction).NotTo(HaveOccurred())
				})

				It("returns true for inserted filename", func() {
					Expect(exists).To(BeTrue())
				})
			})
		})
	})
})
