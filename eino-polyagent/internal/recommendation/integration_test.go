package recommendation

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendationAgentSystem(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	ctx := context.Background()

	t.Run("Complete Recommendation Business Loop", func(t *testing.T) {
		// Create orchestrator
		orchestrator := NewRecommendationOrchestrator(nil, logger)
		require.NotNil(t, orchestrator)

		// 1. Initialize DataAgent
		dataAgent, err := NewDataAgent("data-agent-001", nil, logger)
		require.NoError(t, err)
		require.NotNil(t, dataAgent)

		err = orchestrator.RegisterAgent(dataAgent)
		require.NoError(t, err)

		// 2. Initialize ModelAgent  
		modelAgent, err := NewModelAgent("model-agent-001", nil, logger)
		require.NoError(t, err)
		require.NotNil(t, modelAgent)

		err = orchestrator.RegisterAgent(modelAgent)
		require.NoError(t, err)

		// Verify agents are registered
		agents := orchestrator.GetAgents()
		assert.Len(t, agents, 2)

		// 3. Test Data Collection Task
		dataCollectionTask := &RecommendationTask{
			Type:     TaskDataCollection,
			Priority: PriorityMedium,
			Parameters: map[string]interface{}{
				"collector": "user_behavior",
				"timerange": "last_7_days",
			},
			MaxRetries: 3,
		}

		result, err := orchestrator.ProcessTask(ctx, dataCollectionTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "dataset_id")

		logger.WithFields(logrus.Fields{
			"task_id":      result.TaskID,
			"dataset_id":   result.Data["dataset_id"],
			"record_count": result.Data["record_count"],
		}).Info("Data collection completed")

		// 4. Test Feature Engineering Task
		featureTask := &RecommendationTask{
			Type:     TaskFeatureEngineering,
			Priority: PriorityHigh,
			Parameters: map[string]interface{}{
				"feature_types": []string{"user_profile", "item_similarity"},
			},
			MaxRetries: 2,
		}

		result, err = orchestrator.ProcessTask(ctx, featureTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "features_generated")

		logger.WithFields(logrus.Fields{
			"features_count": result.Data["features_generated"],
			"feature_names":  result.Data["feature_names"],
		}).Info("Feature engineering completed")

		// 5. Test Model Training Task
		trainingTask := &RecommendationTask{
			Type:     TaskModelTraining,
			Priority: PriorityCritical,
			Parameters: map[string]interface{}{
				"algorithm": "collaborative_filtering",
				"hyperparameters": map[string]interface{}{
					"num_factors":    64,
					"learning_rate":  0.01,
					"regularization": 0.001,
				},
			},
			MaxRetries: 1,
		}

		result, err = orchestrator.ProcessTask(ctx, trainingTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "model_id")

		modelID := result.Data["model_id"].(string)
		logger.WithFields(logrus.Fields{
			"model_id":       modelID,
			"algorithm":      result.Data["algorithm"],
			"training_time":  result.Data["training_time"],
		}).Info("Model training completed")

		// 6. Test Model Evaluation Task
		evaluationTask := &RecommendationTask{
			Type:     TaskModelEvaluation,
			Priority: PriorityHigh,
			Parameters: map[string]interface{}{
				"model_id": modelID,
			},
			MaxRetries: 2,
		}

		result, err = orchestrator.ProcessTask(ctx, evaluationTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "evaluation_metrics")

		logger.WithField("evaluation_metrics", result.Data["evaluation_metrics"]).Info("Model evaluation completed")

		// 7. Test Hyperparameter Tuning Task
		tuningTask := &RecommendationTask{
			Type:     TaskHyperParamTuning,
			Priority: PriorityMedium,
			Parameters: map[string]interface{}{
				"algorithm": "collaborative_filtering",
			},
			MaxRetries: 1,
		}

		result, err = orchestrator.ProcessTask(ctx, tuningTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "best_parameters")

		logger.WithFields(logrus.Fields{
			"best_parameters": result.Data["best_parameters"],
			"best_score":      result.Data["best_score"],
		}).Info("Hyperparameter tuning completed")

		// 8. Test Model Deployment Task
		deploymentTask := &RecommendationTask{
			Type:     TaskModelDeployment,
			Priority: PriorityCritical,
			Parameters: map[string]interface{}{
				"model_id": modelID,
			},
			MaxRetries: 1,
		}

		result, err = orchestrator.ProcessTask(ctx, deploymentTask)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "deployment_status")

		logger.WithFields(logrus.Fields{
			"model_id":         modelID,
			"deployment_status": result.Data["deployment_status"],
		}).Info("Model deployment completed")
	})

	t.Run("Agent Performance Metrics", func(t *testing.T) {
		// Create agents for testing
		dataAgent, err := NewDataAgent("perf-data-agent", nil, logger)
		require.NoError(t, err)

		modelAgent, err := NewModelAgent("perf-model-agent", nil, logger)
		require.NoError(t, err)

		// Test agent capabilities
		dataCapabilities := dataAgent.GetCapabilities()
		assert.Contains(t, dataCapabilities, "data_collection")
		assert.Contains(t, dataCapabilities, "feature_engineering")

		modelCapabilities := modelAgent.GetCapabilities()
		assert.Contains(t, modelCapabilities, "model_training")
		assert.Contains(t, modelCapabilities, "hyperparameter_tuning")

		// Test health checks
		assert.NoError(t, dataAgent.HealthCheck())
		assert.NoError(t, modelAgent.HealthCheck())

		// Test performance stats
		dataStats := dataAgent.GetPerformanceStats()
		assert.NotNil(t, dataStats)
		assert.True(t, dataStats.Uptime > 0)

		modelStats := modelAgent.GetPerformanceStats()
		assert.NotNil(t, modelStats)
		assert.True(t, modelStats.Uptime > 0)

		logger.WithFields(logrus.Fields{
			"data_agent_uptime":  dataStats.Uptime,
			"model_agent_uptime": modelStats.Uptime,
		}).Info("Agent performance metrics collected")
	})

	t.Run("Task Priority and Retry Logic", func(t *testing.T) {
		orchestrator := NewRecommendationOrchestrator(nil, logger)
		
		dataAgent, err := NewDataAgent("retry-test-agent", nil, logger)
		require.NoError(t, err)
		
		err = orchestrator.RegisterAgent(dataAgent)
		require.NoError(t, err)

		// Test task with invalid parameters (should fail and retry)
		invalidTask := &RecommendationTask{
			Type:     TaskDataCollection,
			Priority: PriorityHigh,
			Parameters: map[string]interface{}{
				"collector": "non_existent_collector",
			},
			MaxRetries: 2,
		}

		result, err := orchestrator.ProcessTask(ctx, invalidTask)
		// Should fail after retries
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "not found")
		assert.Equal(t, 2, invalidTask.RetryCount)

		logger.WithFields(logrus.Fields{
			"task_id":     result.TaskID,
			"retry_count": invalidTask.RetryCount,
			"error":       result.Error,
		}).Info("Task retry logic tested")

		// Test valid task with higher priority
		validTask := &RecommendationTask{
			Type:     TaskDataValidation,
			Priority: PriorityCritical,
			Parameters: map[string]interface{}{
				"validation_rules": []string{"completeness", "accuracy"},
			},
			MaxRetries: 1,
		}

		result, err = orchestrator.ProcessTask(ctx, validTask)
		require.NoError(t, err)
		assert.True(t, result.Success)

		logger.WithField("validation_result", result.Data).Info("Data validation completed")
	})

	t.Run("System Metrics and Monitoring", func(t *testing.T) {
		orchestrator := NewRecommendationOrchestrator(nil, logger)

		// Register multiple agents
		for i := 0; i < 3; i++ {
			dataAgent, err := NewDataAgent(fmt.Sprintf("metrics-data-agent-%d", i), nil, logger)
			require.NoError(t, err)
			err = orchestrator.RegisterAgent(dataAgent)
			require.NoError(t, err)
		}

		for i := 0; i < 2; i++ {
			modelAgent, err := NewModelAgent(fmt.Sprintf("metrics-model-agent-%d", i), nil, logger)
			require.NoError(t, err)
			err = orchestrator.RegisterAgent(modelAgent)
			require.NoError(t, err)
		}

		// Get system metrics
		systemMetrics := orchestrator.GetSystemMetrics()
		assert.Equal(t, 5, systemMetrics.TotalAgents)
		assert.Equal(t, 5, systemMetrics.ActiveAgents)
		assert.Equal(t, 0, systemMetrics.QueuedTasks)
		assert.Equal(t, 0, systemMetrics.ProcessingTasks)

		logger.WithFields(logrus.Fields{
			"total_agents":   systemMetrics.TotalAgents,
			"active_agents":  systemMetrics.ActiveAgents,
			"queued_tasks":   systemMetrics.QueuedTasks,
		}).Info("System metrics collected")
	})

	t.Run("Agent Configuration Updates", func(t *testing.T) {
		modelAgent, err := NewModelAgent("config-test-agent", nil, logger)
		require.NoError(t, err)

		// Update configuration
		newConfig := map[string]interface{}{
			"max_concurrent_training": 5,
			"auto_deploy_threshold":   0.90,
		}

		err = modelAgent.UpdateConfiguration(newConfig)
		assert.NoError(t, err)
		assert.Equal(t, 5, modelAgent.config.MaxConcurrentTraining)
		assert.Equal(t, 0.90, modelAgent.config.AutoDeployThreshold)

		logger.Info("Agent configuration updated successfully")
	})
}

func TestDataAgentFeatures(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce log noise

	t.Run("Data Collection and Processing", func(t *testing.T) {
		dataAgent, err := NewDataAgent("feature-test-agent", nil, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Test different data collection tasks
		collectors := []string{"user_behavior", "item_catalog", "interaction_logs"}
		
		for _, collector := range collectors {
			task := &RecommendationTask{
				Type: TaskDataCollection,
				Parameters: map[string]interface{}{
					"collector": collector,
				},
			}

			result, err := dataAgent.Process(ctx, task)
			require.NoError(t, err)
			assert.True(t, result.Success)
			assert.NotNil(t, result.Metrics)
			assert.True(t, result.Metrics.ExecutionTime > 0)
		}

		// Verify agent metrics
		metrics := dataAgent.GetMetrics()
		assert.Equal(t, int64(3), metrics.TasksProcessed)
		assert.Equal(t, 1.0, metrics.SuccessRate)
		assert.Equal(t, int64(0), metrics.ErrorCount)
	})

	t.Run("Feature Engineering Pipeline", func(t *testing.T) {
		dataAgent, err := NewDataAgent("feature-pipeline-agent", nil, logger)
		require.NoError(t, err)

		ctx := context.Background()

		featureTypes := [][]string{
			{"user_profile", "demographic_features"},
			{"item_similarity", "content_features"},
			{"interaction_history", "behavioral_features"},
		}

		for _, features := range featureTypes {
			task := &RecommendationTask{
				Type: TaskFeatureEngineering,
				Parameters: map[string]interface{}{
					"feature_types": features,
				},
			}

			result, err := dataAgent.Process(ctx, task)
			require.NoError(t, err)
			assert.True(t, result.Success)
			assert.Equal(t, len(features), result.Data["features_generated"])
		}
	})
}

func TestModelAgentCapabilities(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Algorithm Training and Evaluation", func(t *testing.T) {
		modelAgent, err := NewModelAgent("algo-test-agent", nil, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Test different algorithms
		algorithms := []string{"collaborative_filtering", "content_based", "matrix_factorization", "deep_learning"}

		for _, algorithm := range algorithms {
			// Training task
			trainingTask := &RecommendationTask{
				Type: TaskModelTraining,
				Parameters: map[string]interface{}{
					"algorithm": algorithm,
					"hyperparameters": map[string]interface{}{
						"learning_rate": 0.001,
						"embedding_dim": 32,
					},
				},
			}

			result, err := modelAgent.Process(ctx, trainingTask)
			require.NoError(t, err, "Training failed for algorithm: %s", algorithm)
			assert.True(t, result.Success)

			modelID := result.Data["model_id"].(string)
			assert.NotEmpty(t, modelID)

			// Evaluation task
			evalTask := &RecommendationTask{
				Type: TaskModelEvaluation,
				Parameters: map[string]interface{}{
					"model_id": modelID,
				},
			}

			result, err = modelAgent.Process(ctx, evalTask)
			require.NoError(t, err, "Evaluation failed for algorithm: %s", algorithm)
			assert.True(t, result.Success)
		}

		// Verify trained models
		models := modelAgent.GetModels()
		assert.Len(t, models, 4)

		for _, model := range models {
			assert.Equal(t, ModelStatusTrained, model.Status)
			assert.NotNil(t, model.TrainingMetrics)
			assert.True(t, model.TrainingTime > 0)
		}
	})

	t.Run("Hyperparameter Optimization", func(t *testing.T) {
		modelAgent, err := NewModelAgent("hyperparam-test-agent", nil, logger)
		require.NoError(t, err)

		ctx := context.Background()

		task := &RecommendationTask{
			Type: TaskHyperParamTuning,
			Parameters: map[string]interface{}{
				"algorithm": "collaborative_filtering",
			},
		}

		result, err := modelAgent.Process(ctx, task)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "best_parameters")
		assert.Contains(t, result.Data, "best_score")

		bestParams := result.Data["best_parameters"].(map[string]interface{})
		assert.Contains(t, bestParams, "learning_rate")
		assert.Contains(t, bestParams, "regularization")

		bestScore := result.Data["best_score"].(float64)
		assert.True(t, bestScore > 0 && bestScore <= 1.0)
	})
}

// Benchmark tests for performance validation
func BenchmarkDataCollection(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during benchmarking

	dataAgent, _ := NewDataAgent("benchmark-data-agent", nil, logger)
	ctx := context.Background()

	task := &RecommendationTask{
		Type: TaskDataCollection,
		Parameters: map[string]interface{}{
			"collector": "user_behavior",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dataAgent.Process(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkModelTraining(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	modelAgent, _ := NewModelAgent("benchmark-model-agent", nil, logger)
	ctx := context.Background()

	task := &RecommendationTask{
		Type: TaskModelTraining,
		Parameters: map[string]interface{}{
			"algorithm": "collaborative_filtering",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := modelAgent.Process(ctx, task)
		if err != nil {
			b.Fatal(err)
		}
	}
}