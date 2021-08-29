package global

import "go.uber.org/zap"

type Any map[string]interface{}

var Logger *zap.SugaredLogger