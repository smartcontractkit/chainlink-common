package grafana

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"gopkg.in/yaml.v3"
)

func NewNotificationTemplatesFromFile(filePath string) (map[string]alerting.NotificationTemplate, error) {
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	notificationTemplates := make(map[string]alerting.NotificationTemplate)
	yamlFileToMapRes, errFileToMap := yamlFileToMap(filePath)
	if errFileToMap != nil {
		return nil, errFileToMap
	}

	for typeTemplate, template := range yamlFileToMapRes {
		newTemplate, err := alerting.NewNotificationTemplateBuilder().
			Name(fmt.Sprintf("%s-%s-notification-template", fileName, typeTemplate)).
			Template(template).Build()
		if err != nil {
			return nil, err
		}
		notificationTemplates[typeTemplate] = newTemplate
	}

	return notificationTemplates, nil
}

func yamlFileToMap(filepath string) (map[string]string, error) {
	yamlFile, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	config := make(map[string]string)

	errUnmarshal := yaml.Unmarshal(yamlFile, &config)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	return config, nil
}
