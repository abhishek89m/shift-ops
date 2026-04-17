package main

func (s *store) getSummary() (summary, error) {
	tasks, err := s.listTasks()
	if err != nil {
		return summary{}, err
	}

	var result summary
	result.Total = len(tasks)

	for _, task := range tasks {
		switch task.Status {
		case statusPending:
			result.Pending++
		case statusInProgress:
			result.InProgress++
		case statusCompleted:
			result.Completed++
		case statusSkipped:
			result.Skipped++
		}
	}

	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM task_events
		WHERE date(created_at) = date('now');
	`).Scan(&result.EventsToday); err != nil {
		return summary{}, err
	}

	result.RecommendedTask = chooseRecommendation(tasks)
	return result, nil
}

func chooseRecommendation(tasks []task) *recommendation {
	bestScore := -1 << 30
	var bestTask task
	var bestReasons []string
	found := false

	for _, task := range tasks {
		if task.Status != statusPending && task.Status != statusInProgress {
			continue
		}

		score, reasons := scoreTask(task)
		if !found || score > bestScore {
			bestScore = score
			bestTask = task
			bestReasons = reasons
			found = true
		}
	}

	if !found {
		return nil
	}

	return &recommendation{
		Task:    bestTask,
		Reasons: bestReasons,
	}
}

func scoreTask(task task) (int, []string) {
	score := task.Urgency * 100
	reasons := make([]string, 0, 4)

	if task.Status == statusInProgress {
		score += 220
		reasons = append(reasons, "Continue active task")
	}

	if task.Urgency >= 3 {
		reasons = append(reasons, "High urgency")
	}

	if task.BlockedAccessSeverity > 0 {
		score += task.BlockedAccessSeverity * 35
		reasons = append(reasons, "Blocked access risk")
	}

	if task.Type == "battery_swap" && task.BatteryLevel != nil && *task.BatteryLevel <= 20 {
		score += 70
		reasons = append(reasons, "Low battery")
	}

	distancePenalty := min(task.DistanceMeters, 1200) / 12
	score -= distancePenalty
	if task.DistanceMeters <= 250 {
		reasons = append(reasons, "Close by")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Next best available task")
	}

	return score, reasons
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
