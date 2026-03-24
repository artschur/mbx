package template

import (
	"context"
	"mbx/models"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTemplateService_GetTemplates_UberMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := NewMockTemplateFetcher(ctrl)
	mockSender := NewMockTemplateSender(ctrl)
	mockRepo := NewMockTemplateRepo(ctrl)

	service := NewTemplateService(mockFetcher, mockSender, mockRepo)

	ctx := context.Background()
	templates := []models.SavedTemplate{
		{ContentId: "1", FriendlyName: "T1"},
	}

	// Set expectations
	mockFetcher.EXPECT().GetTemplates(ctx).Return(templates, nil)
	mockRepo.EXPECT().SaveTemplate(ctx, templates[0]).Return(nil)

	res, err := service.GetTemplates(ctx)

	assert.NoError(t, err)
	assert.Equal(t, templates, res)
}

func TestTemplateService_CreateTemplate_UberMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFetcher := NewMockTemplateFetcher(ctrl)
	mockSender := NewMockTemplateSender(ctrl)
	mockRepo := NewMockTemplateRepo(ctrl)

	service := NewTemplateService(mockFetcher, mockSender, mockRepo)

	ctx := context.Background()
	dto := models.CreateTemplateDTO{FriendlyName: "New"}
	expected := &models.SavedTemplate{ContentId: "2", FriendlyName: "New"}

	mockSender.EXPECT().CreateTemplate(ctx, dto).Return(expected, nil)
	mockRepo.EXPECT().SaveTemplate(ctx, *expected).Return(nil)

	res, err := service.CreateTemplate(ctx, dto)

	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}
