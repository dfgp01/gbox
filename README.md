
#   预期目标

需要用分層的思路去實現，分層的下一步是分塊，先把所有邏輯做好獨立切割，形成各自獨立，再根據層級關係形成調用路徑，這樣才能避免一直重複設計和推翻，要先從大體上進行模糊抽象，現盡可能列出層級和部分模塊

* gbox項目會定義和覆蓋這些層


#	lib層

	説明：這一層的所有部分都應該是原子化的，各部不應有依賴關係，這是重要前提
		1、約定格式和接口行爲，這些底層機制將影響整體的運作
		2、約定接口行爲
		3、提供默認封裝，可以引入第三方庫，但是要挑選經過認證的
	舉例：
		定義序列化行爲：Marshal() []byte, error
		默認json實現：json.Marshal()、pb.Marshal
		這樣，既可以提供默認的實現，也可以給外部自行擴展

原子模塊|説明
:---:|:---:
序列化|提供json和pb默認，擴展[]pb或map[]pb這些格式（已有）
反射|這個是重頭，提供迭代定義和默認實現（開發中）
組件|定義組件、事件，全局管理器等機制
數據報|Msg部分，如msgId、seq、body等字段内容，錯誤碼，還有消息管理器
狀態機|定義狀態機的機制行爲，如外部管理和内部節點，提供默認實現（事件和單綫機制，過期處理等）
算法|雪花、一致性哈希、時間輪等，這部分不能滲入外部庫，除非官方高認可度
基礎工具|貨幣計算，其他數值計算，字符串處理等，基礎環境，如本地ipv4、env參數等，可滲入外部第三方庫
MTL|主要是日志接口，度量和追蹤主要用於微服務


# 驅動層

  説明：這裏主要的功能是封裝所有client組件的邏輯，依賴lib庫

原子模塊|説明
:---:|:---:
client組件|mysql, redis, mongo, etcd, kafka, grpc
orm層|包裝mysql、redis、mongo等的反射，依賴client組件和lib.reflector
通訊層|ws、http的邏輯，主要是client部分，依賴lib.message
MTL層|主要是log的封裝，trace和metric的理解暫時不多，擴展例子：zap到kafka的組件封裝


# 伺服層
  説明：凡是需要啓動時開綫程處理的都算服務層組件，如grpc-server、redis-brpop、kafka-consumer等，默認實現會依賴MTL和其他下游組件

原子模塊|説明|componentType
:---:|:---:|:---:
http-server|封裝gin|HttpServer
ws-server|封裝ws|WsServer
web-server|聯合http和ws|WebServer
rpc-server|裝grpc|RPCServer
tcp, udp|同理|TcpServer, UdpServer
  
	未確定：
		kafka-consumer，在component層的基礎上，封裝一個goroutine的service(start, stop)，提供業務接口
		redis-bpop


# 業務層
  説明：service組件，這裏主要提供組裝業務組件的平臺，自由組裝多個server和其他組件，根據并行機制，制定自己業務


# 通訊層
	説明：包含http、tcp、udp、ws、grpc，主要依賴封裝的msg結構，因爲要做統一的格式，grpc可能會特殊處理，也可能不在通訊層内，也可以是默認一個grpc服務函數，是處理msg信息的





# 拆解模塊開發任務（不分先後）

###	事件

	1、管理器定義接口Send(evt)、AddListener(evtType, handler)、RemoveListener(evtType)、Clear()
	2、默認提供全局事件管理器
	3、事件定義統一的結構，不一定要復用msg報文，所有上層組件都需要遵守，如默認提供的etcd.watch()發生時，在内部也轉爲事件消息報並用事件管理器發送，然後調用自定義的ListenerHandler
	4、問題：在整個運行周期裏，如何調度時間管理器？還是由業務自己決定？

###	組件
	貫徹整個gbox的意志，就是統一的組件和事件處理機制，提供統一的開發標準流程
	
	1、組件根據id和type定義，如mysqlCom的id=1001，type=rdb
	2、組件管理器是全局唯一的，僅在啓動和關閉時對map進行寫，不用加鎖和原子化
	3、組件管理器根據getById和getByType()查找組件，因此，兩屬性均有唯一性，其中以type查找居多
	4、若想同時啓用多個rdb組件，需自行在外叠加一層，同樣type=rdb，自行處理邏輯即可

###	日志：
	1、Info()等標準接口
	2、caller調用顯示、最好可以自行序列化，提供外部接口
	3、調研zap和logrus實現