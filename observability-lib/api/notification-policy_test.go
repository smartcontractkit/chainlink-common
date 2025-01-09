package api

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/stretchr/testify/require"
)

func Pointer[T any](d T) *T {
	return &d
}

func TestObjectMatchersEqual(t *testing.T) {
	t.Run("returns true if the two object matchers are equal", func(t *testing.T) {
		a := alerting.ObjectMatchers{{"team", "=", "chainlink"}}
		b := alerting.ObjectMatchers{{"team", "=", "chainlink"}}

		result := objectMatchersEqual(a, b)
		require.True(t, result)
	})

	t.Run("returns true if the two object matchers with multiple matches are equal", func(t *testing.T) {
		a := alerting.ObjectMatchers{
			{"team", "=", "chainlink"},
			{"severity", "=", "critical"},
		}
		b := alerting.ObjectMatchers{
			{"severity", "=", "critical"},
			{"team", "=", "chainlink"},
		}

		result := objectMatchersEqual(a, b)
		require.True(t, result)
	})

	t.Run("returns false if the two object matchers with multiple matches are different", func(t *testing.T) {
		a := alerting.ObjectMatchers{
			{"team", "=", "chainlink"},
			{"severity", "=", "critical"},
		}
		b := alerting.ObjectMatchers{
			{"severity", "=", "warning"},
			{"team", "=", "chainlink"},
		}

		result := objectMatchersEqual(a, b)
		require.False(t, result)
	})
}

func TestPolicyExists(t *testing.T) {
	t.Run("policyExists return true if policy exists", func(t *testing.T) {
		notificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
					Routes: []alerting.NotificationPolicy{
						{
							Receiver: Pointer("pagerduty"),
							ObjectMatchers: &alerting.ObjectMatchers{
								{"env", "=", "production"},
							},
						},
					},
				},
			},
		}

		newNotificationPolicy := alerting.NotificationPolicy{
			Receiver: Pointer("pagerduty"),
			ObjectMatchers: &alerting.ObjectMatchers{
				{"env", "=", "production"},
			},
		}
		result := policyExist(*notificationPolicyTree, newNotificationPolicy)
		require.True(t, result)
	})

	t.Run("policyExists return false if policy does not exists", func(t *testing.T) {
		notificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
					Routes: []alerting.NotificationPolicy{
						{
							Receiver: Pointer("pagerduty"),
							ObjectMatchers: &alerting.ObjectMatchers{
								{"env", "=", "production"},
							},
						},
					},
				},
			},
		}

		newNotificationPolicy := alerting.NotificationPolicy{
			Receiver: Pointer("pagerduty"),
			ObjectMatchers: &alerting.ObjectMatchers{
				{"key", "=", "value"},
			},
		}
		result := policyExist(*notificationPolicyTree, newNotificationPolicy)
		require.False(t, result)
	})

	t.Run("updateInPlace should update notification policy if already exists", func(t *testing.T) {
		notificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
					Routes: []alerting.NotificationPolicy{
						{
							Receiver: Pointer("pagerduty"),
							ObjectMatchers: &alerting.ObjectMatchers{
								{"env", "=", "production"},
							},
						},
					},
				},
			},
		}

		newNotificationPolicy := alerting.NotificationPolicy{
			Receiver: Pointer("pagerduty"),
			ObjectMatchers: &alerting.ObjectMatchers{
				{"env", "=", "production"},
			},
			Continue: Pointer(true),
		}

		expectedNotificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
					Routes: []alerting.NotificationPolicy{
						{
							Receiver: Pointer("pagerduty"),
							ObjectMatchers: &alerting.ObjectMatchers{
								{"env", "=", "production"},
							},
							Continue: Pointer(true),
						},
					},
				},
			},
		}

		updateInPlace(notificationPolicyTree, newNotificationPolicy)
		require.Equal(t, expectedNotificationPolicyTree, notificationPolicyTree)
	})

	t.Run("deleteInPlace should delete notification policy if exists", func(t *testing.T) {
		notificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
				},
				{
					Receiver: Pointer("slack2"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink2"},
					},
				},
				{
					Receiver: Pointer("slack3"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink3"},
					},
				},
			},
		}

		newNotificationPolicy := alerting.NotificationPolicy{
			Receiver: Pointer("slack2"),
			ObjectMatchers: &alerting.ObjectMatchers{
				{"team", "=", "chainlink2"},
			},
		}

		expectedNotificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
				},
				{
					Receiver: Pointer("slack3"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink3"},
					},
				},
			},
		}
		deleteInPlace(notificationPolicyTree, newNotificationPolicy)
		require.Equal(t, expectedNotificationPolicyTree, notificationPolicyTree)
	})

	t.Run("deleteInPlace should delete notification policy if exists", func(t *testing.T) {
		notificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
					Routes: []alerting.NotificationPolicy{
						{
							Receiver: Pointer("pagerduty"),
							ObjectMatchers: &alerting.ObjectMatchers{
								{"env", "=", "production"},
							},
						},
					},
				},
			},
		}

		newNotificationPolicy := alerting.NotificationPolicy{
			Receiver: Pointer("pagerduty"),
			ObjectMatchers: &alerting.ObjectMatchers{
				{"env", "=", "production"},
			},
		}

		expectedNotificationPolicyTree := &alerting.NotificationPolicy{
			Receiver: Pointer("grafana-default-email"),
			Routes: []alerting.NotificationPolicy{
				{
					Receiver: Pointer("slack"),
					ObjectMatchers: &alerting.ObjectMatchers{
						{"team", "=", "chainlink"},
					},
				},
			},
		}
		deleteInPlace(notificationPolicyTree, newNotificationPolicy)
		require.Equal(t, expectedNotificationPolicyTree, notificationPolicyTree)
	})

}
