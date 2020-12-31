package log

import (
	"bytes"
	"fmt"
	json "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"io"
	"log"
	"math"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

//@description	日志记录工具类
/*
日志格式:	时间(系统时间)	日志类型(方法设置)		日志内容(动态输入)
日志类包含两个同步锁:	缓冲区锁-mu_buf	文件锁-mu_file
	日志输入操作	Printf	Println
				1.获取缓冲区锁
				2.写入缓冲区
				3.释放缓冲区锁
				4.A.调用bufWrite,B.等待定时调用bufWrite
	日志输出操作	bufWrite
				1.获取文件锁
				2.判断缓冲区,不需写入则返回
				3.获取缓冲区锁
				4.写入缓冲区
				5.释放缓冲区锁
	日志监听操作	fileMonitor
				A.文件监听定时器到期fileCheck
					1.判断是否需要文件重名,并后续操作
					1.1.获取文件锁
					1.2.再次判断文件是否需要重名
					1.3.重命名文件
					1.4.释放文件锁
				B.定时写入定时器到期bufWrite

	文件定时写入bufWrite与文件监听fileMonitor时间间隔 t1,t2
	防止文件碰撞(秒为单位时)需要满足	(n-1)t1%60 != (n-1)t2%60
	顺序获取锁:缓冲锁-->文件锁
*/
//@author hanse
//@data 2016-08-04	13:39	调试代码无报错
//		2017-04-08	17:04	修改bufWrite获取锁机制,写入方式采用线程实现,增加必要的设计说明
//		2017-05-24	21:11	代码结构调整,合并非不需方法,对检测及写入事件进行动态设置,并对定时器进行重置

var Logg *Logger

const (
	_VERSION_          = "1.0.1"               //版本
	DATEFORMAT         = "2006-01-02"          //日期格式(用于文件命名)
	TIMEFORMAT         = "2006/01/02 15:04:05" //时间格式(日志时间格式)
	_SPACE             = " "                   //参数分割
	_TABLE             = "\t"                  //日志文件行分隔符
	_JOIN              = "&"                   //参数连接符
	_FILE_OPERAT_MODE_ = 0644                  //文件操作权限模式
	_FILE_CREAT_MODE_  = 0666                  //文件建立权限模式
	_LABEL_            = "[_loggor_]"          //标签
)
const (
	//日志文件存储模式
	LOG_FILE_SAVE_USUAL = 1 //普通模式,不分割
	LOG_FILE_SAVE_SIZE  = 2 //大小分割
	LOG_FILE_SAVE_DATE  = 3 //日期分割
)
const (
	//文件大小单位
	_        = iota
	KB int64 = 1 << (iota * 10)
	MB
	GB
	TB
)
const (
	_EXTEN_NAME_               = ".log"                  //日志文件后缀名
	_CHECK_TIME_ time.Duration = 900 * time.Millisecond  //定时检测是否分割检测周期
	_WRITE_TIME_ time.Duration = 1300 * time.Millisecond //定时写入文件周期
)

var (
	IS_DEBUG     = false //调试模式
	TIMEER_WRITE = false //定时写入文件
)

type LOGGER interface {
	SetDebug(bool)                                      //设置日志文件路径及名称
	SetType(uint)                                       //设置日志类型
	SetRollingFile(string, string, int32, int64, int64) //按照文件大小分割
	SetRollingDaily(string, string)                     //按照日期分割
	SetRollingNormal(string, string)                    //设置普通模式
	Close()                                             //关闭
	Println(a ...interface{})                           //打印日志
	Printf(format string, a ...interface{})             //格式化输出
}

//==================================================================日志记录器
type Logger struct {
	log_type     uint          //日志类型
	path         string        //日志文件路径
	dir          string        //目录
	filename     string        //文件名
	maxFileSize  int64         //文件大小
	maxFileCount int32         //文件个数
	dailyRolling bool          //日分割
	sizeRolling  bool          //大小分割
	nomalRolling bool          //普通模式(不分割)
	_suffix      int           //大小分割文件的当前序号
	_date        *time.Time    //文件时间
	mu_buf       *sync.Mutex   //缓冲锁
	mu_file      *sync.Mutex   //文件锁
	logfile      *os.File      //文件句柄
	timer        *time.Timer   //监视定时器
	writeTimer   *time.Timer   //批量写入定时器
	buf          *bytes.Buffer //缓冲区(公用buf保证数据写入的顺序性)
}

/**获取日志对象**/
func New() *Logger {

	l := &Logger{}
	l.buf = &bytes.Buffer{}
	l.mu_buf = new(sync.Mutex)
	l.mu_file = new(sync.Mutex)
	return l
}

//error
func Error(a ...interface{}) {
	tp := "[error]" + fmt.Sprint(a...)
	handle(2, tp)
}
func Errorf(format string, a ...interface{}) {
	tp := "[error]" + fmt.Sprintf(format, a...)
	handle(2, tp)
}

/**格式行输出**/
func Printf(format string, a ...interface{}) {
	tp := "[info]" + fmt.Sprintf(format, a...)
	handle(1, tp)
}

/**逐行输出**/
func Println(a ...interface{}) {
	tp := "[info]" + fmt.Sprint(a...)
	handle(1, tp)
}

// pType 1 info 2 error
func colorPrint(pType int, tp string) {
	_, file, line, _ := runtime.Caller(3)
	//funcName:=runtime.FuncForPC(funcNamePtr).Name()
	var logType string
	var color string
	if pType == 1 {
		logType = "info"
		color = "1;35;1"
	} else if pType == 2 {
		logType = "eroor"
		color = "1;31;1"
	}
	fmt.Printf("\n%c[%s m[%s][%s](%s)\n%s%c[0m\n", 0x1B, color, logType, time.Now().Format(TIMEFORMAT), file+":"+cast.ToString(line), tp, 0x1B)
}

func handle(pType int, tp string) {
	defer func() {
		if !TIMEER_WRITE {
			go bufWrite()
		}
	}()

	Logg.mu_buf.Lock()
	colorPrint(pType, tp)
	defer Logg.mu_buf.Unlock()
	Logg.buf.WriteString(
		fmt.Sprintf(
			"%s\t%d\t%s\n",
			time.Now().Format(TIMEFORMAT),
			Logg.log_type,
			tp,
		),
	)
}

/**测试模式**/
func SetDebug(is_debug bool) {
	IS_DEBUG = is_debug
}

/**定时写入**/
func SetTimeWrite(time_write bool) *Logger {
	TIMEER_WRITE = time_write

	return Logg
}

/**日志类型**/
func SetType(tp uint) {
	Logg.log_type = tp
}

/**大小分割**/
func SetRollingFile(dir, _file string, maxn int32, maxs int64, _u int64) {
	//0.输入合法性
	if Logg.sizeRolling ||
		Logg.dailyRolling ||
		Logg.nomalRolling {
		log.Println(_LABEL_, "mode can't be changed!")
		return
	}

	//1.设置各模式标志符
	Logg.sizeRolling = true
	Logg.dailyRolling = false
	Logg.nomalRolling = false

	//2.设置日志器各参数
	Logg.maxFileCount = maxn
	Logg.maxFileSize = maxs * int64(_u)
	Logg.dir = dir
	Logg.filename = _file
	for i := 1; i <= int(maxn); i++ {
		sizeFile := fmt.Sprintf(
			dir,
			_file,
			_EXTEN_NAME_,
			".",
			fmt.Sprintf("%05d", i),
		)
		if isExist(sizeFile) {
			Logg._suffix = i
		} else {
			break
		}
	}
	//3.实时文件写入
	Logg.path = fmt.Sprint(
		dir,
		_file,
		_EXTEN_NAME_,
	)
	startLogger(Logg.path)
}

/**日期分割**/
func SetRollingDaily(dir, _file string) {
	//0.输入合法性
	if Logg.sizeRolling ||
		Logg.dailyRolling ||
		Logg.nomalRolling {
		log.Println(_LABEL_, "mode can't be changed!")
		return
	}

	//1.设置各模式标志符
	Logg.sizeRolling = false
	Logg.dailyRolling = true
	Logg.nomalRolling = false

	//2.设置日志器各参数
	Logg.dir = dir
	Logg.filename = _file
	Logg._date = getNowFormDate(DATEFORMAT)
	startLogger(
		fmt.Sprint(
			Logg.dir,
			Logg.filename,
			Logg._date.Format(DATEFORMAT),
			_EXTEN_NAME_,
		),
	)
}

/**普通模式**/
func SetRollingNormal(dir, _file string) {
	//0.输入合法性
	if Logg.sizeRolling ||
		Logg.dailyRolling ||
		Logg.nomalRolling {
		log.Println(_LABEL_, "mode can't be changed!")
		return
	}

	//1.设置各模式标志符
	Logg.sizeRolling = false
	Logg.dailyRolling = false
	Logg.nomalRolling = true

	//2.设置日志器各参数
	Logg.dir = dir
	Logg.filename = _file
	startLogger(
		fmt.Sprint(
			dir,
			_file,
			_EXTEN_NAME_,
		),
	)
}

/**关闭日志器**/
func Close() {
	//0.获取锁
	Logg.mu_buf.Lock()
	defer Logg.mu_buf.Unlock()
	Logg.mu_file.Lock()
	defer Logg.mu_file.Unlock()

	//1.关闭
	if nil != Logg.timer {
		Logg.timer.Stop()
	}
	if nil != Logg.writeTimer {
		Logg.writeTimer.Stop()
	}
	if Logg.logfile != nil {
		err := Logg.logfile.Close()

		if err != nil {
			log.Println(_LABEL_, "file close err", err)
		}
	} else {
		log.Println(_LABEL_, "file has been closed!")
	}

	//2.清理
	Logg.sizeRolling = false
	Logg.dailyRolling = false
	Logg.nomalRolling = false
}

//==================================================================内部工具方法
//初始日志记录器(各日志器统一调用)
func startLogger(tp string) {
	defer func() {
		if e, ok := recover().(error); ok {
			log.Println(_LABEL_, "WARN: panic - %v", e)
			log.Println(_LABEL_, string(debug.Stack()))
		}
	}()

	//1.初始化空间
	var err error
	Logg.buf = &bytes.Buffer{}
	Logg.mu_buf = new(sync.Mutex)
	Logg.mu_file = new(sync.Mutex)
	Logg.path = tp
	checkFileDir(tp)
	Logg.logfile, err = os.OpenFile(
		tp,
		os.O_RDWR|os.O_APPEND|os.O_CREATE,
		_FILE_OPERAT_MODE_,
	)
	if nil != err {
		log.Println(_LABEL_, "OpenFile err!")
	}

	//2.开启监控线程
	go func() {
		Logg.timer = time.NewTimer(_CHECK_TIME_)
		Logg.writeTimer = time.NewTimer(_WRITE_TIME_)
		if !TIMEER_WRITE {
			Logg.writeTimer.Stop()
		}

		for {
			select {
			//定时检测是否分割
			case <-Logg.timer.C:
				fileCheck()
				if IS_DEBUG && false {
					log.Printf("*") //心跳
				}
				break
			//定时写入文件(定时写入,会导致延时)
			case <-Logg.writeTimer.C:
				bufWrite()
				if IS_DEBUG && false {
					log.Printf(".") //心跳
				}
				break
			}
		}
	}()

	if IS_DEBUG {
		jstr, err := json.Marshal(Logg)
		if nil == err {
			log.Println(_LABEL_, _VERSION_, string(jstr))
		}
	}
}

//文件检测(会锁定文件)
func fileCheck() {
	//0.边界判断
	if nil == Logg.mu_file ||
		nil == Logg.logfile ||
		"" == Logg.path {

		return
	}
	defer func() {
		if e, ok := recover().(error); ok {
			log.Println(_LABEL_, "WARN: panic - %v", e)
			log.Println(_LABEL_, string(debug.Stack()))
		}
	}()

	//1.重命名判断
	var RENAME_FLAG bool = false
	var CHECK_TIME time.Duration = _CHECK_TIME_
	Logg.timer.Stop()
	defer Logg.timer.Reset(CHECK_TIME)
	if Logg.dailyRolling {
		//日分割模式
		now := getNowFormDate(DATEFORMAT)
		if nil != now &&
			nil != Logg._date &&
			now.After(*Logg._date) {
			//超时重名
			RENAME_FLAG = true
		} else {
			//检测间隔动态调整
			du := Logg._date.UnixNano() - now.UnixNano()
			abs := math.Abs(float64(du))
			CHECK_TIME = CHECK_TIME * time.Duration(abs/abs)
		}
	} else if Logg.sizeRolling {
		//文件大小模式
		if "" != Logg.path &&
			Logg.maxFileCount >= 1 &&
			fileSize(Logg.path) >= Logg.maxFileSize {
			//超量重名
			RENAME_FLAG = true
		}
	} else if Logg.nomalRolling {
		//普通模式
		RENAME_FLAG = false
	}

	//2.重名操作
	if RENAME_FLAG {
		Logg.mu_file.Lock()
		defer Logg.mu_file.Unlock()
		if IS_DEBUG {
			log.Println(_LABEL_, Logg.path, "is need rename.")
		}
		fileRename()
	}

	return
}

//重命名文件
func fileRename() {
	//1.生成文件名称
	var err error
	var newName string
	var oldName string
	defer func() {
		if IS_DEBUG {
			log.Println(
				_LABEL_,
				oldName,
				"->",
				newName,
				":",
				err,
			)
		}
	}()

	if Logg.dailyRolling {
		//日期分割模式(文件不重命名)
		oldName = Logg.path
		newName = Logg.path
		Logg._date = getNowFormDate(DATEFORMAT)
		Logg.path = fmt.Sprint(
			Logg.dir,
			Logg.filename,
			Logg._date.Format(DATEFORMAT),
			_EXTEN_NAME_,
		)
	} else if Logg.sizeRolling {
		//大小分割模式(1,2,3....)
		suffix := int(Logg._suffix%int(Logg.maxFileCount) + 1)
		oldName = Logg.path
		newName = fmt.Sprint(
			Logg.path,
			".",
			fmt.Sprintf("%05d", suffix),
		)
		Logg._suffix = suffix
		Logg.path = Logg.path
	} else if Logg.nomalRolling {
		//常规模式
	}

	//2.处理旧文件
	Logg.logfile.Close()
	if "" != oldName && "" != newName && oldName != newName {
		if isExist(newName) {
			//删除旧文件
			err := os.Remove(newName)
			if nil != err {
				log.Println(_LABEL_, "remove file err", err.Error())
			}
		}
		err = os.Rename(oldName, newName)
		if err != nil {
			//重名旧文件
			log.Println(_LABEL_, "rename file err", err.Error())
		}
	}

	//3.创建新文件
	Logg.logfile, err = os.OpenFile(
		Logg.path,
		os.O_RDWR|os.O_APPEND|os.O_CREATE,
		_FILE_OPERAT_MODE_,
	)
	if err != nil {
		log.Println(_LABEL_, "creat file err", err.Error())
	}

	return
}

//缓冲写入文件
func bufWrite() {
	//0.边界处理
	if nil == Logg.buf ||
		"" == Logg.path ||
		nil == Logg.logfile ||
		nil == Logg.mu_buf ||
		nil == Logg.mu_file ||
		Logg.buf.Len() <= 0 {
		return
	}

	//1.数据写入
	var WRITE_TIME time.Duration = _WRITE_TIME_
	if nil != Logg.writeTimer {
		Logg.writeTimer.Stop()
		defer Logg.writeTimer.Reset(WRITE_TIME)
	}
	Logg.mu_file.Lock()
	defer Logg.mu_file.Unlock()
	Logg.mu_buf.Lock()
	defer Logg.mu_buf.Unlock()
	defer Logg.buf.Reset()
	n, err := io.WriteString(Logg.logfile, Logg.buf.String())
	if nil != err {
		//写入失败,校验文件,不存在则创建
		checkFileDir(Logg.path)
		Logg.logfile, err = os.OpenFile(
			Logg.path,
			os.O_RDWR|os.O_APPEND|os.O_CREATE,
			_FILE_OPERAT_MODE_,
		)
		if nil != err {
			log.Println(_LABEL_, "log bufWrite() err!")
		}
	}
	//根据缓冲压力进行动态设置写入间隔
	if n == 0 {
		WRITE_TIME = _WRITE_TIME_
	} else {
		WRITE_TIME = WRITE_TIME * time.Duration(n/n)
	}
}

//==================================================================辅助方法
//获取文件大小
func fileSize(file string) int64 {
	Logg, e := os.Stat(file)
	if e != nil {
		if IS_DEBUG {
			log.Println(_LABEL_, e.Error())
		}
		return 0
	}

	return Logg.Size()
}

//判断路径是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)

	return err == nil || os.IsExist(err)
}

//检查文件路径文件夹,不存在则创建
func checkFileDir(tp string) {
	p, _ := path.Split(tp)
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(p, _FILE_CREAT_MODE_); err != nil {
			log.Println(_LABEL_, "CheckFileDir() Creat dir faile!")
		}
	}
}

//获取当前指定格式的日期
func getNowFormDate(form string) *time.Time {
	t, err := time.Parse(form, time.Now().Format(form))
	if nil != err {
		log.Println(_LABEL_, "getNowFormDate()", err.Error())
		t = time.Time{}

		return &t
	}

	return &t
}

func init() {
	Logg = New()
	SetType(LOG_FILE_SAVE_DATE)
	SetRollingDaily("./logs/", "log")
}
