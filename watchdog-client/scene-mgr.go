package main

import "fmt"

type SceneManager struct {
	scenes map[string]string
}

func InitSceneManager() SceneManager {
	return SceneManager{scenes: make(map[string]string)}
}

func (sm *SceneManager) InjectScene(key string, value string) {
	sm.scenes[key] = value
}

func (sm *SceneManager) GetScene(key string) (value string, ok bool) {
	if value, ok = sm.scenes[key]; ok {
		return value, true
	}

	return "", false
}

func (sm *SceneManager) ToString() string {
	result := ""
	for key, value := range sm.scenes {
		result += fmt.Sprintf("SCENE | %s = %s\n", key, value)
	}
	return result
}
