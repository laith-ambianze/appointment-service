# Task: Return Partial Failure Details from BulkCreateAvailabilityRules

**Status:** 🔴 TODO  
**Priority:** High (API correctness)  
**Files:** `internal/service/availability_service.go`, `internal/handlers/availability.go`  
**Function:** `BulkCreateAvailabilityRules`, `BulkCreateAvailabilityRules` handler

---

## Objective

`BulkCreateAvailabilityRules` currently silently swallows every per-rule error and always returns `nil` as the error — meaning the caller receives `HTTP 201` even when every rule failed to create. Add a structured partial-failure response so callers can detect and act on failures.

---

## Current State (Bug)

```go
// availability_service.go
func (s *AvailabilityService) BulkCreateAvailabilityRules(ctx context.Context, productID uuid.UUID, rules []CreateAvailabilityRuleRequest) ([]models.AvailabilityRule, error) {
    createdRules := make([]models.AvailabilityRule, 0, len(rules))

    for _, req := range rules {
        rule, err := s.CreateAvailabilityRule(ctx, productID, req)
        if err != nil {
            s.logger.Warn("Failed to create availability rule", ...)
            continue  // error silently discarded
        }
        createdRules = append(createdRules, *rule)
    }

    return createdRules, nil  // always nil — caller cannot detect failures
}
```

The handler returns `HTTP 201` unconditionally:

```go
// handlers/availability.go
c.JSON(http.StatusCreated, gin.H{
    "provider_id": providerID,
    "created":     rules,
    "count":       len(rules),
})
```

---

## Target State

Return a structured result distinguishing succeeded and failed items. Use `HTTP 207 Multi-Status` when there is a mix of successes and failures. Use `HTTP 201` only when all rules succeeded. Use `HTTP 400` when all rules failed.

---

## Implementation Steps

### 1. Add result types to `internal/service/availability_service.go`

```go
// BulkCreateRuleResult holds the outcome for a single rule in a bulk request
type BulkCreateRuleResult struct {
    DayOfWeek int                    `json:"day_of_week"`
    ProviderID string                `json:"provider_id"`
    Rule      *models.AvailabilityRule `json:"rule,omitempty"`
    Error     string                 `json:"error,omitempty"`
    Success   bool                   `json:"success"`
}

// BulkCreateAvailabilityRulesResult holds the full outcome of a bulk create
type BulkCreateAvailabilityRulesResult struct {
    Results       []BulkCreateRuleResult   `json:"results"`
    SucceededCount int                     `json:"succeeded_count"`
    FailedCount    int                     `json:"failed_count"`
}
```

### 2. Update `BulkCreateAvailabilityRules` signature and logic

```go
func (s *AvailabilityService) BulkCreateAvailabilityRules(ctx context.Context, productID uuid.UUID, rules []CreateAvailabilityRuleRequest) (*BulkCreateAvailabilityRulesResult, error) {
    result := &BulkCreateAvailabilityRulesResult{
        Results: make([]BulkCreateRuleResult, 0, len(rules)),
    }

    for _, req := range rules {
        item := BulkCreateRuleResult{
            DayOfWeek:  req.DayOfWeek,
            ProviderID: req.ProviderID,
        }

        rule, err := s.CreateAvailabilityRule(ctx, productID, req)
        if err != nil {
            s.logger.Warn("Failed to create availability rule in bulk",
                zap.String("provider_id", req.ProviderID),
                zap.Int("day_of_week", req.DayOfWeek),
                zap.Error(err),
            )
            item.Error = err.Error()
            item.Success = false
            result.FailedCount++
        } else {
            item.Rule = rule
            item.Success = true
            result.SucceededCount++
        }

        result.Results = append(result.Results, item)
    }

    return result, nil
}
```

### 3. Update the handler in `internal/handlers/availability.go`

```go
func (h *AvailabilityHandler) BulkCreateAvailabilityRules(c *gin.Context) {
    // ... existing binding logic unchanged ...

    result, err := h.service.BulkCreateAvailabilityRules(c.Request.Context(), productID, rules)
    if err != nil {
        h.handleServiceError(c, err)
        return
    }

    // Determine HTTP status based on outcome
    status := http.StatusCreated
    if result.SucceededCount == 0 {
        status = http.StatusBadRequest
    } else if result.FailedCount > 0 {
        status = http.StatusMultiStatus // 207
    }

    c.JSON(status, gin.H{
        "provider_id":     providerID,
        "results":         result.Results,
        "succeeded_count": result.SucceededCount,
        "failed_count":    result.FailedCount,
    })
}
```

---

## Acceptance Criteria

- [ ] All rules succeed → `HTTP 201`, `failed_count = 0`.
- [ ] Some rules fail → `HTTP 207`, both `succeeded_count` and `failed_count` are non-zero, each result item has `success`, `day_of_week`, and either `rule` or `error`.
- [ ] All rules fail → `HTTP 400`, `succeeded_count = 0`.
- [ ] Per-item `error` field contains a human-readable message (not an internal stack trace).
- [ ] Unit test covers the mixed-success scenario.
