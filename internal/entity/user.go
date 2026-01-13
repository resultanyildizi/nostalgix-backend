package entity

// User represents a user.
type User struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Credits      int           `json:"credits"`
	Subscription *Subscription `json:"subscription"`
	AuthMethod   string        `json:"auth_method"`
	IsNewUser    bool          `json:"is_new_user"`
	CustomerID   string        `json:"customer_id"`
	FCMToken     string        `json:"-"`
	AuthID       string        `json:"-"`
}

type SubscriptionType string

const (
	SubscriptionTypeTrial   SubscriptionType = "trial"
	SubscriptionTypeIntro   SubscriptionType = "intro"
	SubscriptionTypeNormal  SubscriptionType = "normal"
	SubscriptionTypePrepaid SubscriptionType = "prepaid"
	SubscriptionTypePromo   SubscriptionType = "promo"
)

type SubscriptionPlan string

const (
	SubscriptionPlanPro SubscriptionType = "pro"
)

type SubscriptionPlanPeriod string

const (
	SubscriptionPlanPeriod1W SubscriptionPlanPeriod = "1w"
	SubscriptionPlanPeriod1M SubscriptionPlanPeriod = "1m"
	SubscriptionPlanPeriod6M SubscriptionPlanPeriod = "6m"
	SubscriptionPlanPeriod1Y SubscriptionPlanPeriod = "1y"
)

func (s SubscriptionPlanPeriod) GetDays() int {
	switch s {
	case SubscriptionPlanPeriod1M:
		return 30
	case SubscriptionPlanPeriod6M:
		return 30 * 6
	case SubscriptionPlanPeriod1W:
		return 7
	case SubscriptionPlanPeriod1Y:
		return 30 * 12
	default:
		return 0
	}
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive       SubscriptionStatus = "active"
	SubscriptionStatusExpired      SubscriptionStatus = "expired"
	SubscriptionStatusBillingIssue SubscriptionStatus = "billing_issue"
)

type Subscription struct {
	Plan   string `json:"plan"`
	Type   string `json:"type"`
	Period string `json:"period"`
	Status string `json:"status"`
}

// GetID returns the user ID.
func (u User) GetID() string {
	return u.ID
}

// GetName returns the user name.
func (u User) GetName() string {
	return u.Name
}
