package pool

// Pool 一个盲盒池：定义（manifest）与可用的支付资源（resources）。
// 池文件本身就是这份结构，resources 缺省时回退到索引里的全局资源文件。
type Pool struct {
	Manifest  Manifest   `json:"manifest"`
	Resources []Resource `json:"resources"`
}

// applyGlobalResources 让没写 resources 的池回退到全局资源。
func applyGlobalResources(item *Pool, global []Resource) {
	if item.Resources == nil {
		item.Resources = global
	}
}
