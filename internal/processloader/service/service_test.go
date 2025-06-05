package service_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/servicefakes"
)

var (
	ErrRead      = errors.New("read error")
	ErrGetName   = errors.New("get name error")
	ErrStoreMeta = errors.New("store meta error")
)

var _ = Describe("Service", func() {
	When("created", func() {
		It("should instantiate the service", func() {
			Expect(service.NewService(nil, nil)).NotTo(BeNil())
		})
	})

	Describe("IndexProcessDefinitions", Serial, func() {
		var (
			svc        *service.Service
			ctx        context.Context
			errAction  error
			fakeStore  *servicefakes.FakeStore
			fakeReader *servicefakes.FakeReader
		)

		BeforeEach(func() {
			ctx = context.Background()
			fakeStore = &servicefakes.FakeStore{}
			fakeReader = &servicefakes.FakeReader{}

			svc = service.NewService(fakeStore, fakeReader)

			fakeReader.ReadYAMLFilesReturns([]string{"dir/process.yaml"}, nil)
			fakeReader.GetProcessNameFromFileReturns("test-process", nil)
			fakeStore.SaveProcessDefinitionMetaReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.IndexProcessDefinitions(ctx)
		})

		It("succeeds when all steps pass", func() {
			Expect(errAction).NotTo(HaveOccurred())

			Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
			Expect(fakeReader.GetProcessNameFromFileCallCount()).To(Equal(1))
			Expect(fakeStore.SaveProcessDefinitionMetaCallCount()).To(Equal(1))

			gotPath := fakeReader.GetProcessNameFromFileArgsForCall(0)
			Expect(gotPath).To(Equal("dir/process.yaml"))

			gotCtx, gotName, gotMetaPath := fakeStore.SaveProcessDefinitionMetaArgsForCall(0)
			Expect(gotCtx).To(Equal(ctx))
			Expect(gotName).To(Equal("test-process"))
			Expect(gotMetaPath).To(Equal("dir/process.yaml"))
		})

		When("reading YAML files fails", func() {
			BeforeEach(func() {
				fakeReader.ReadYAMLFilesReturns(nil, ErrRead)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrRead))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeReader.GetProcessNameFromFileCallCount()).To(Equal(0))
				Expect(fakeStore.SaveProcessDefinitionMetaCallCount()).To(Equal(0))
			})
		})

		When("getting process name fails", func() {
			BeforeEach(func() {
				fakeReader.GetProcessNameFromFileReturns("", ErrGetName)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrGetName))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeReader.GetProcessNameFromFileCallCount()).To(Equal(1))
				Expect(fakeStore.SaveProcessDefinitionMetaCallCount()).To(Equal(0))
			})
		})

		When("saving metadata fails", func() {
			BeforeEach(func() {
				fakeStore.SaveProcessDefinitionMetaReturns(ErrStoreMeta)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrStoreMeta))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeReader.GetProcessNameFromFileCallCount()).To(Equal(1))
				Expect(fakeStore.SaveProcessDefinitionMetaCallCount()).To(Equal(1))
			})
		})

		When("multiple files are returned", func() {
			BeforeEach(func() {
				fakeReader.ReadYAMLFilesReturns([]string{"file1.yaml", "file2.yaml"}, nil)
			})

			It("handles all files", func() {
				Expect(errAction).NotTo(HaveOccurred())
				Expect(fakeReader.GetProcessNameFromFileCallCount()).To(Equal(2))
				Expect(fakeStore.SaveProcessDefinitionMetaCallCount()).To(Equal(2))

				for i := 0; i < 2; i++ {
					path := fakeReader.GetProcessNameFromFileArgsForCall(i)
					Expect(path).To(ContainSubstring(".yaml"))

					_, name, metaPath := fakeStore.SaveProcessDefinitionMetaArgsForCall(i)
					Expect(name).To(Equal("test-process"))
					Expect(metaPath).To(ContainSubstring(".yaml"))
				}
			})
		})
	})
})
