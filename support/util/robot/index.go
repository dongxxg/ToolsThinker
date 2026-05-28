/**
 * @author  zhaoliang.liang
 * @date  2024/12/16 13:39
 */

package robot

import (
	"fmt"
	"os"
	"strings"
)

func GetRobotCloudPushJobName(env, businessId string) string {
	if len(businessId) > 38 {
		businessId = businessId[:38]
	}
	jobName := fmt.Sprintf(
		"%s-live-%s-%s",
		strings.ToLower(env),
		getShortEnv(),
		strings.ReplaceAll(businessId, "_", "-"),
	)
	return jobName
}

func getShortEnv() string {
	env := os.Getenv("APP_ENV")
	switch strings.ToLower(env) {
	case "production", "prod":
		return "p"
	case "development", "dev":
		return "d"
	case "staging", "stage":
		return "s"
	case "test":
		return "t"
	default:
		return "d"
	}
}
