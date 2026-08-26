package retention

import "time"

func EligibleAt(receivedAt time.Time, policy Policy) time.Time {
	return receivedAt.UTC().Add(policy.KeepFor + policy.Grace)
}

func EligibleObjects(now time.Time, policy Policy, objects []Object) []Object {
	eligible := objects[:0]
	for _, object := range objects {
		if !now.Before(EligibleAt(object.ReceivedAt, policy)) {
			eligible = append(eligible, object)
		}
	}
	return eligible
}
