package agollo

import (
	"github.com/zouyx/agollo/agcache"
	"strconv"
	"sync"
	"sync/atomic"
)

//ConfigFileFormat 配置文件类型
type ConfigFileFormat string

const (
	//Properties
	Properties ConfigFileFormat = "properties"
	//XML
	XML ConfigFileFormat = "xml"
	//JSON
	JSON ConfigFileFormat = "json"
	//YML
	YML ConfigFileFormat = "yml"
	//YAML
	YAML ConfigFileFormat = "yaml"
)

const (
	empty = ""

	//1 minute
	configCacheExpireTime = 120

	defaultNamespace  = "application"
	defaultContentKey = "content"
)

var (
	currentConnApolloConfig = &currentApolloConfig{
		configs: make(map[string]*ApolloConnConfig, 1),
	}

	//config from apollo
	apolloConfigCache = make(map[string]*Config, 0)

	formatParser        = make(map[ConfigFileFormat]ContentParser, 0)
	defaultFormatParser = &DefaultParser{}

	cacheFactory = &agcache.DefaultCacheFactory{}
)

func init() {
	formatParser[Properties] = &PropertiesParser{}
}

func initDefaultConfig() {
	initConfigCache(cacheFactory)
}

//initNamespaceConfig 根据namespace创建缓存
func initNamespaceConfig(namespace string) {

	createNamespaceConfig(cacheFactory, namespace)

	initNamespaceNotifications(namespace)
}

func initConfigCache(cacheFactory *agcache.DefaultCacheFactory) {
	if appConfig == nil {
		logger.Warn("Config is nil,can not init agollo.")
		return
	}
	createNamespaceConfig(cacheFactory, appConfig.NamespaceName)
}

func createNamespaceConfig(cacheFactory *agcache.DefaultCacheFactory, namespace string) {
	splitNamespaces(namespace, func(namespace string) {
		if apolloConfigCache[namespace] != nil {
			return
		}
		c := &Config{
			namespace: namespace,
			cache:     cacheFactory.Create(),
		}
		c.isInit.Store(false)
		c.waitInit.Add(1)
		apolloConfigCache[namespace] = c
	})
}

type currentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*ApolloConnConfig
}

//Config apollo配置项
type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

//getIsInit 获取标志
func (this *Config) getIsInit() bool {
	return this.isInit.Load().(bool)
}

//getConfigValue 获取配置值
func (this *Config) getConfigValue(key string) interface{} {
	b := this.getIsInit()
	if !b {
		this.waitInit.Wait()
	}
	if this.cache == nil {
		logger.Errorf("get config value fail!namespace:%s is not exist!", this.namespace)
		return empty
	}

	value, err := this.cache.Get(key)
	if err != nil {
		logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return empty
	}

	return string(value)
}

//getValue 获取配置值（string）
func (this *Config) getValue(key string) string {
	value := this.getConfigValue(key)
	if value == nil {
		return empty
	}

	return value.(string)
}

//GetStringValue 获取配置值（string），获取不到则取默认值
func (this *Config) GetStringValue(key string, defaultValue string) string {
	value := this.getValue(key)
	if value == empty {
		return defaultValue
	}

	return value
}

//GetIntValue 获取配置值（int），获取不到则取默认值
func (this *Config) GetIntValue(key string, defaultValue int) int {
	value := this.getValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

//GetFloatValue 获取配置值（float），获取不到则取默认值
func (this *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := this.getValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

//GetBoolValue 获取配置值（bool），获取不到则取默认值
func (this *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := this.getValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

//GetConfig 根据namespace获取apollo配置
func GetConfig(namespace string) *Config {
	return GetConfigAndInit(namespace)
}

//GetConfigAndInit 根据namespace获取apollo配置
func GetConfigAndInit(namespace string) *Config {
	if namespace == "" {
		return nil
	}

	if apolloConfigCache[namespace] == nil {
		initNamespaceConfig(namespace)

		notifySimpleSyncConfigServices(namespace)
	}
	return apolloConfigCache[namespace]
}

//GetConfigCache 根据namespace获取apollo配置的缓存
func GetConfigCache(namespace string) agcache.CacheInterface {
	config := GetConfigAndInit(namespace)
	if config == nil {
		return nil
	}
	if !config.getIsInit() {
		config.waitInit.Wait()
	}

	return config.cache
}

func getDefaultConfigCache() agcache.CacheInterface {
	config := GetConfigAndInit(defaultNamespace)
	if config != nil {
		return config.cache
	}
	return nil
}

func updateApolloConfig(apolloConfig *ApolloConfig, isBackupConfig bool) {
	if apolloConfig == nil {
		logger.Error("apolloConfig is null,can't update!")
		return
	}
	//get change list
	changeList := updateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime, apolloConfig.NamespaceName)

	if len(changeList) > 0 {
		//create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		//push change event to channel
		pushChangeEvent(event)
	}

	//update apollo connection config
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.configs[apolloConfig.NamespaceName] = &apolloConfig.ApolloConnConfig

	if isBackupConfig {
		//write config file async
		go writeConfigFile(apolloConfig, appConfig.getBackupConfigPath())
	}
}

func updateApolloConfigCache(configurations map[string]string, expireTime int, namespace string) map[string]*ConfigChange {
	config := GetConfig(namespace)
	if config == nil {
		return nil
	}

	isInit := false
	defer func(c *Config) {
		if !isInit {
			return
		}
		b := c.getIsInit()
		if b {
			return
		}
		c.isInit.Store(isInit)
		c.waitInit.Done()
	}(config)

	if (configurations == nil || len(configurations) == 0) && config.cache.EntryCount() == 0 {
		return nil
	}

	//get old keys
	mp := map[string]bool{}
	config.cache.Range(func(key, value interface{}) bool {
		mp[key.(string)] = true
		return true
	})

	changes := make(map[string]*ConfigChange)

	if configurations != nil {
		// update new
		// keys
		for key, value := range configurations {
			//key state insert or update
			//insert
			if !mp[key] {
				changes[key] = createAddConfigChange(value)
			} else {
				//update
				oldValue, _ := config.cache.Get(key)
				if string(oldValue) != value {
					changes[key] = createModifyConfigChange(string(oldValue), value)
				}
			}

			config.cache.Set(key, []byte(value), expireTime)
			delete(mp, string(key))
		}
	}

	// remove del keys
	for key := range mp {
		//get old value and del
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(string(oldValue))

		config.cache.Del(key)
	}
	isInit = true

	return changes
}

//base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace: nameSpace,
		Changes:   changes,
	}
}

func touchApolloConfigCache() error {
	return nil
}

//GetApolloConfigCache 获取默认namespace的apollo配置
func GetApolloConfigCache() agcache.CacheInterface {
	return getDefaultConfigCache()
}

//GetCurrentApolloConfig 获取Apollo链接配置
func GetCurrentApolloConfig() map[string]*ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.configs
}

func getCurrentApolloConfigReleaseKey(namespace string) string {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.configs[namespace]
	if config == nil {
		return empty
	}

	return config.ReleaseKey
}

func getConfigValue(key string) interface{} {
	value, err := getDefaultConfigCache().Get(key)
	if err != nil {
		logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return empty
	}

	return string(value)
}

func getValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return empty
	}

	return value.(string)
}

func GetStringValue(key string, defaultValue string) string {
	value := getValue(key)
	if value == empty {
		return defaultValue
	}

	return value
}

func GetIntValue(key string, defaultValue int) int {
	value := getValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string, defaultValue float64) float64 {
	value := getValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string, defaultValue bool) bool {
	value := getValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

//GetContent 获取配置文件内容
func (c *Config) GetContent(format ConfigFileFormat) string {
	parser := formatParser[format]
	if parser == nil {
		parser = defaultFormatParser
	}
	s, err := parser.parse(c.cache)
	if err != nil {
		logger.Debug("GetContent fail ! error:", err)
	}
	return s
}
