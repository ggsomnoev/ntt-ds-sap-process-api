package store_test

import (
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
			name      string
			path      string
		)

		BeforeEach(func() {
			s = store.NewStore(pool)

			name = "sample-process"
			path = "/etc/config/sample.yaml"
		})

		Describe("SaveProcessDefinitionMeta", func() {
			JustBeforeEach(func() {
				errAction = s.SaveProcessDefinitionMeta(ctx, name, path)
			})

			It("succeeds", func() {
				Expect(errAction).NotTo(HaveOccurred())
			})

			It("saves a new process definition entry", func() {
				storedPath, err := s.GetProcessPathByName(ctx, name)
				Expect(err).NotTo(HaveOccurred())

				Expect(storedPath).To(Equal(path))
			})

			It("updates the path on conflict", func() {
				newPath := "/new/location/updated.yaml"
				Expect(s.SaveProcessDefinitionMeta(ctx, name, newPath)).To(Succeed())

				var storedPath string

				storedPath, err := s.GetProcessPathByName(ctx, name)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedPath).To(Equal(newPath))
			})
		})

		Describe("GetProcessPathByName", func() {
			var storedPath string
			BeforeEach(func() {
				Expect(s.SaveProcessDefinitionMeta(ctx, name, path)).To(Succeed())
			})

			JustBeforeEach(func() {
				storedPath, errAction = s.GetProcessPathByName(ctx, name)
			})

			It("succeeds", func() {
				Expect(errAction).NotTo(HaveOccurred())
			})

			It("returns the correct path for an existing name", func() {
				Expect(storedPath).To(Equal(path))
			})

			It("returns an error if the name doesn't exist", func() {
				_, err := s.GetProcessPathByName(ctx, "nonexistent")
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})
	})
})
