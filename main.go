package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Dreamacro/clash/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
)

func main() {
	// using standard library "flag" package
	port := *flag.String("port", "8080", "default  8080")
	path := *flag.String("path", "rewrite", "default  rewrite")

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./")
	viper.ReadInConfig()

	// 设置路由，如果访问/，则调用index方法
	http.HandleFunc("/"+path, index)

	// 启动web服务，监听9090端口
	log.Infoln(fmt.Sprintf("server start port:%s , path:%s ", port, path))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	config.Init()
	//ParseProxy
}

// w表示response对象，返回给客户端的内容都在对象里处理
// r表示客户端请求对象，包含了请求头，请求参数等等
func index(w http.ResponseWriter, r *http.Request) {
	reqToken := r.URL.Query().Get("token")
	token := viper.GetString("token")
	if reqToken != token {
		fmt.Fprintf(w, "Hello golang http!")
		return
	}

	yaml := conversionYaml()
	// 往w里写入内容，就会在浏览器里输出
	fmt.Fprintf(w, string(yaml))
}

func conversionYaml() []byte {
	url := viper.GetString("url")
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Error(err)
	}
	all, err := io.ReadAll(resp.Body)
	//log.Println(string(all))

	config := new(ClashX)
	yaml.Unmarshal(all, config)

	stringMap := viper.Get("proxy")
	sliceProxy, err := ToSliceProxy(stringMap)
	log.Infoln(sliceProxy)
	//.GetStringMap("proxy")
	marshal, _ := json.Marshal(stringMap)
	proxy := new([]Proxy)
	yaml.Unmarshal(marshal, proxy)

	proxys := make([]Proxy, 0)
	proxys = append(proxys, *proxy...)
	proxys = append(proxys, config.Proxies...)
	config.Proxies = proxys

	//追加到组
	groups := make([]ProxyGroup, 0)
	for _, group := range config.ProxyGroups {
		tmpNames := make([]string, 0)
		for _, p := range *proxy {
			tmpNames = append(tmpNames, p.Name)
		}
		tmpNames = append(tmpNames, group.Proxies...)
		group.Proxies = tmpNames

		groups = append(groups, group)
	}
	config.ProxyGroups = groups
	out, _ := yaml.Marshal(config)
	return out
}

func ToSliceProxy(i interface{}) ([]Proxy, error) {
	var a []Proxy

	switch v := i.(type) {
	case []interface{}:
		for _, proxy := range v {
			p := new(Proxy)
			out, _ := yaml.Marshal(proxy)
			yaml.Unmarshal(out, p)
			a = append(a, *p)
		}
		return a, nil
	default:
		return nil, fmt.Errorf("unable to cast %#v of type %T to map[string]interface{}", i, i)
	}
}

type ClashX struct {
	Port               int64             `yaml:"port"`
	SocksPort          int64             `yaml:"socks-port"`
	RedirPort          int64             `yaml:"redir-port"`
	MixedPort          int64             `yaml:"mixed-port"`
	AllowLAN           bool              `yaml:"allow-lan"`
	Mode               string            `yaml:"mode"`
	LogLevel           string            `yaml:"log-level"`
	Ipv6               bool              `yaml:"ipv6"`
	Hosts              map[string]string `yaml:"hosts"`
	ExternalController string            `yaml:"external-controller"`
	ClashForAndroid    ClashForAndroid   `yaml:"clash-for-android"`
	Profile            Profile           `yaml:"profile"`
	DNS                DNS               `yaml:"dns"`
	Proxies            []Proxy           `yaml:"proxies"`
	ProxyGroups        []ProxyGroup      `yaml:"proxy-groups"`
	Rules              []string          `yaml:"rules"`
}

type ClashForAndroid struct {
	AppendSystemDNS bool `yaml:"append-system-dns"`
}

type DNS struct {
	Enable            bool           `yaml:"enable"`
	Listen            string         `yaml:"listen"`
	DefaultNameserver []string       `yaml:"default-nameserver"`
	Ipv6              bool           `yaml:"ipv6"`
	EnhancedMode      string         `yaml:"enhanced-mode"`
	FakeIPFilter      []string       `yaml:"fake-ip-filter"`
	Nameserver        []string       `yaml:"nameserver"`
	Fallback          []string       `yaml:"fallback"`
	FallbackFilter    FallbackFilter `yaml:"fallback-filter"`
}

type FallbackFilter struct {
	Geoip  bool     `yaml:"geoip"`
	Ipcidr []string `yaml:"ipcidr"`
	Domain []string `yaml:"domain"`
}

//type Hosts struct {
//	ServicesGoogleapisCN string `yaml:"services.googleapis.cn"`
//	WWWGoogleCN          string `yaml:"www.google.cn"`
//}

type Hosts struct {
	ServicesGoogleapisCN string `yaml:"services.googleapis.cn"`
	WWWGoogleCN          string `yaml:"www.google.cn"`
}

type Profile struct {
	Tracing bool `yaml:"tracing"`
}

type Proxy struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Server     string                 `yaml:"server"`
	Port       int64                  `yaml:"port"`
	Cipher     string                 `yaml:"cipher"`
	Password   string                 `yaml:"password,omitempty"`
	Uuid       string                 `yaml:"uuid,omitempty"`
	AlterId    string                 `yaml:"alterId,omitempty"`
	UDP        bool                   `yaml:"udp,omitempty"`
	Tls        bool                   `yaml:"tls,omitempty"`
	Network    string                 `yaml:"network,omitempty"`
	WsPath     string                 `yaml:"ws-path,omitempty"`
	Plugin     string                 `yaml:"plugin,omitempty"`
	PluginOpts map[string]interface{} `yaml:"plugin-opts,omitempty"`
}

type ProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      *string  `yaml:"url,omitempty"`
	Interval *int64   `yaml:"interval,omitempty"`
}
