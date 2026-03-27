package adapter

import "fmt"

type Factory struct {
	adapters map[string]Adapter
}

func NewFactory() *Factory {
	f := &Factory{
		adapters: make(map[string]Adapter),
	}
	f.registerDefaultAdapters()
	return f
}

func (f *Factory) registerDefaultAdapters() {
	f.Register("openai", NewOpenAIAdapter())
	f.Register("nvidia", NewNVIDIAAdapter())
	f.Register("nvidia_nim", NewNVIDIAAdapter())
	f.Register("azure", NewAzureAdapter())
	f.Register("claude", NewClaudeAdapter())
	f.Register("anthropic", NewClaudeAdapter())
	f.Register("gemini", NewGeminiAdapter())
	f.Register("google", NewGeminiAdapter())
	f.Register("deepseek", NewDeepSeekAdapter())
	f.Register("zhipu", NewZhipuAdapter())
	f.Register("baidu", NewBaiduAdapter())
	f.Register("yi", NewYiAdapter())
	f.Register("yiapi", NewYiAPIAdapter())
	f.Register("ollama", NewOllamaAdapter())
	f.Register("localai", NewLocalAIAdapter())
	f.Register("groq", NewGroqAdapter())
}

func (f *Factory) Register(channelType string, adapter Adapter) {
	f.adapters[channelType] = adapter
}

func (f *Factory) Get(channelType string) (Adapter, error) {
	adapter, ok := f.adapters[channelType]
	if !ok {
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
	return adapter, nil
}

func (f *Factory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.adapters))
	for t := range f.adapters {
		types = append(types, t)
	}
	return types
}

func (f *Factory) GetAdapterName(channelType string) string {
	adapter, err := f.Get(channelType)
	if err != nil {
		return channelType
	}
	return adapter.GetName()
}

var defaultFactory *Factory

func GetFactory() *Factory {
	if defaultFactory == nil {
		defaultFactory = NewFactory()
	}
	return defaultFactory
}

func GetAdapter(channelType string) (Adapter, error) {
	return GetFactory().Get(channelType)
}

func SupportedTypes() []string {
	return GetFactory().GetSupportedTypes()
}
