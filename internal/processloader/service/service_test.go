package service_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/servicefakes"
)

var (
	ErrRead         = errors.New("read error")
	ErrParse        = errors.New("parse error")
	ErrValidation   = errors.New("invalid process")
	ErrSend         = errors.New("send failure")
	ErrWrite        = errors.New("db write failed")
	ErrMarkComplete = errors.New("mark complete error")
)

var _ = Describe("Service", func() {
	When("created", func() {
		It("exists", func() {
			Expect(service.NewService(nil, nil, nil, nil)).NotTo(BeNil())
		})
	})

	var _ = Describe("instance", Serial, func() {
		var (
			svc           *service.Service
			ctx           context.Context
			errAction     error
			fakeStore     *servicefakes.FakeStore
			fakeReader    *servicefakes.FakeReader
			fakeValidator *servicefakes.FakeValidator
			fakeSender    *servicefakes.FakeSender
		)

		BeforeEach(func() {
			ctx = context.Background()
			fakeStore = &servicefakes.FakeStore{}
			fakeReader = &servicefakes.FakeReader{}
			fakeValidator = &servicefakes.FakeValidator{}
			fakeSender = &servicefakes.FakeSender{}

			svc = service.NewService(fakeStore, fakeReader, fakeValidator, fakeSender)

			fakeStore.RunInAtomicallyStub = func(_ context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}

			fakeReader.ReadYAMLFilesReturns([]string{"dir/process.yaml"}, nil)
			fakeStore.FileExistsReturns(false, nil)
			fakeStore.AddProcessedFileReturns(nil)
			fakeReader.ParseConfigFileReturns(model.ProcessDefinition{Name: "test"}, nil)
			fakeValidator.ValidateReturns(nil)
			fakeSender.SendReturns(nil)
			fakeStore.MarkCompletedReturns(nil)
		})

		JustBeforeEach(func() {
			errAction = svc.TryProcessConfigs(ctx)
		})

		It("succeeds", func() {
			Expect(errAction).NotTo(HaveOccurred())

			Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
			Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
			Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
			Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(1))
			Expect(fakeValidator.ValidateCallCount()).To(Equal(1))
			Expect(fakeSender.SendCallCount()).To(Equal(1))
			Expect(fakeStore.MarkCompletedCallCount()).To(Equal(1))
		})

		When("reading YAML files fails", func() {
			BeforeEach(func() {
				fakeReader.ReadYAMLFilesReturns(nil, ErrRead)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrRead))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(0))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(0))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(0))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(0))
				Expect(fakeSender.SendCallCount()).To(Equal(0))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("processed file already exists", func() {
			BeforeEach(func() {
				fakeStore.FileExistsReturns(true, nil)
			})

			It("succeeds", func() {
				Expect(errAction).NotTo(HaveOccurred())

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(0))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(0))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(0))
				Expect(fakeSender.SendCallCount()).To(Equal(0))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("adding to processed file fails", func() {
			BeforeEach(func() {
				fakeStore.AddProcessedFileReturns(ErrWrite)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrWrite))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(0))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(0))
				Expect(fakeSender.SendCallCount()).To(Equal(0))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("parsing config fails", func() {
			BeforeEach(func() {
				fakeReader.ParseConfigFileReturns(model.ProcessDefinition{}, ErrParse)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrParse))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(1))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(0))
				Expect(fakeSender.SendCallCount()).To(Equal(0))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("validation fails", func() {
			BeforeEach(func() {
				fakeValidator.ValidateReturns(ErrValidation)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrValidation))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(1))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(1))
				Expect(fakeSender.SendCallCount()).To(Equal(0))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("send fails", func() {
			BeforeEach(func() {
				fakeSender.SendReturns(ErrSend)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrSend))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(1))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(1))
				Expect(fakeSender.SendCallCount()).To(Equal(1))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(0))
			})
		})

		When("marking file as complete fails", func() {
			BeforeEach(func() {
				fakeStore.MarkCompletedReturns(ErrMarkComplete)
			})

			It("returns an error", func() {
				Expect(errAction).To(MatchError(ErrMarkComplete))

				Expect(fakeReader.ReadYAMLFilesCallCount()).To(Equal(1))
				Expect(fakeStore.FileExistsCallCount()).To(Equal(1))
				Expect(fakeStore.AddProcessedFileCallCount()).To(Equal(1))
				Expect(fakeReader.ParseConfigFileCallCount()).To(Equal(1))
				Expect(fakeValidator.ValidateCallCount()).To(Equal(1))
				Expect(fakeSender.SendCallCount()).To(Equal(1))
				Expect(fakeStore.MarkCompletedCallCount()).To(Equal(1))
			})
		})

		// TODO: cover mulptiple config files
	})
})
