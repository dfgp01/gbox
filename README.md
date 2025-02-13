
# 预期目标

先進行抽象分層，然後是逐層的分塊，所有邏輯形成各自獨立（高内聚，低耦合），根據層級關係形成調用路徑，這樣才能避免一直重複設計和推翻，要先從大體上進行模糊抽象，盡可能列出層級和部分模塊

* gbox將定義和覆蓋這些層


# 原子模型層(lib)

	預期：
		1、這一層的所有部分都應該是原子化的，互不依賴，這是重要前提
		2、只定義最基礎的機制，不參雜任何第三方（算法除外）
		3、約定格式和接口行爲，這些底層機制將影響整體的運作
		4、提供默認封裝，必要時可引入第三方庫
	舉例：
		定義序列化行爲：Marshal() []byte, error
		默認json實現：json.Marshal()、pb.Marshal
		這樣，既可以提供默認的實現，也可以給外部自行擴展

模塊|説明
:---:|:---:
事件|初版已完成，目前只是同步串行listener
組件|由RW兩個綫程組成，至少運行一個綫程，發送事件和監聽都由異步執行，提供option機制，所有綫程入口都要recover()，注意交叉chan死鎖問題
消息|僅定義報文格式，可以封裝[]pb，提供序列化：json和proto；預設msg模型：code, data, message等結構
反射|提供迭代遍歷機制，未完成
MTL|定義公共日志接口、度量、追蹤等，僅定義；日志格式可配置、抽象、公共化，度量和追蹤最好也支持
幾何|一些數學運算和定律的函數封裝


# 工具庫

	預期：
		1、封裝其他第三方，一些好用的工具實現
		2、可獨立使用，不需其他依賴的
	舉例：
		1、JWT
		2、rest-client

模塊|説明
:---:|:---:
日志|zap, logrus(擱置), std, rotate, lumber 進行交叉封裝，實現定義好的日志接口
認證|JWT、oauth2等
追蹤、度量|tracer、metric，暫不明確
基礎工具|貨幣計算，其他數值計算，字符串處理等，基礎環境，如本地ipv4、env參數等，可用第三方庫
算法|雪花、一致性哈希、時間輪等，盡量不引用複雜的三方實現
狀態機|需要依賴事件


# 驅動層

	預期：
		1、封裝client使用爲主
		2、依賴lib或util層
		3、基本上只針對發送端
	舉例：
		1、mysql, redis, mongo, etcd, kafka等
		2、grpc-client

原子模塊|説明
:---:|:---:
client組件|mysql, redis, mongo, etcd, kafka, grpc
orm層|包裝mysql、redis、mongo等的反射，依賴client組件和lib.reflector
通訊層|ws、http的邏輯，主要是client部分，依賴lib.message
MTL層|主要是log的封裝，trace和metric的理解暫時不多，擴展例子：zap到kafka的組件封裝


# 伺服層

	預期：
		1、凡是需要啓動時開綫程處理的都算服務層組件
		2、默認實現會依賴MTL和其他下游組件
		3、每個協議端口對應一個server-id和addr
		4、只封裝服務，不耦合log或jwt等其他無關元素，可在業務層再組合
	舉例：
		1、grpc-server、http-server等
		2、redis-sub
		3、kafka-consumer

原子模塊|説明|componentType
:---:|:---:|:---:
http-server|封裝gin|HttpServer
ws-server|封裝ws|WsServer
web-server|聯合http和ws|WebServer
rpc-server|裝grpc|RPCServer
tcp, udp|同理|TcpServer, UdpServer


# 業務層

	預期：
		1、暴露給外部使用
		2、提供自己的業務邏輯
		3、組裝server、client、util、lib.message等多個層級的組件
	舉例：
		1、IAP模塊
		2、分佈式鎖


# 更高層


場景|説明
:---:|:---:
報文|統一http、tcp、udp、ws的消息格式，可能無法兼容grpc，可提供默認的grpc實現，用於處理統一msg的



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