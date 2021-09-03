package goms

import "encoding/json"

func RemoteTrace(rc *ReqContext) {
	logJson, _ := json.Marshal(rc)
	_, err := Es.Put("logs", string(logJson))
	if err != nil {
		Logger.Errorf("[elasticsearch] %s", err)
	}
}
