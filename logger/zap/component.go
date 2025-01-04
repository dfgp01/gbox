package zp

//1.封裝zapLogger，公共接口，還有zap的一些函數轉換，因爲zapLogger應該不和component耦合
//2.ZapComponent包裝zapLogger，用組件機制
//3.提供zapLogger和zapComponent出去，給到外部選擇

//use zap Option

type ZapComponent struct {
	//base component
	driver *ZapLogger
}

func NewZapComponent() *ZapComponent {
	return nil
}
