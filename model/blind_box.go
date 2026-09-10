package model

// BlindBox 一个盲盒：它的定义（Manifest）以及可用的支付货币。
type BlindBox struct {
	Manifest   Manifest
	Currencies []Currency
}
