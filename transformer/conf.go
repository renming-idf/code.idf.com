package transformer

type Conf struct {
	Mysql    Mysql
	App      App
	TestData TestData
	Data     Data
}

type App struct {
	Name          string
	URl           string
	Port          string
	CertFile      string
	KeyFile       string
	LoggerLevel   string
	DirverType    string
	CreateSysData bool
}

type Mysql struct {
	DirverName string
	Connect    string
	Name       string
	TName      string
	CasbinName string
}

type TestData struct {
	UserName string
	Name     string
	Pwd      string
}
type Data struct {
	SecretPassphrase string
	KeyStoreDir      string
	Uploads          string
	MainWalletsPath  string
	Url              string
	MyUrl            string
}
