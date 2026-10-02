package worker

import (
	"context"
	"fmt"
)

type Task func(ctx context.Context, config map[string]any) error;

func placeholderTask(ctx context.Context, config map[string]any) error {
	if shouldFail, ok := config["simulate_failure"].(bool); ok && shouldFail {
		return fmt.Errorf("simulated task failure");
	}

	return nil;
}