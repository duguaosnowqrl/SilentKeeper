#定义环境
ifeq ($(OS),Windows_NT)
    # Windows 用 cmd 的 set
    GOENV_LINUX = set GOOS=linux&& set GOARCH=amd64&&
    GOENV_WIN64 = set GOOS=windows&& set GOARCH=amd64&&
    GOENV_WIN32 = set GOOS=windows&& set GOARCH=386&&
    RM = del /Q
    RMDIR = rmdir /S /Q
    MKDIR = mkdir
    COPY = copy /Y
    COPY_DIR = xcopy /Y /E /I
    fix_path = $(subst /,\,$(1))
else
    # Linux/Mac 用 Unix 语法
    GOENV_LINUX = GOOS=linux GOARCH=amd64
    GOENV_WIN64 = GOOS=windows GOARCH=amd64
    GOENV_WIN32 = GOOS=windows GOARCH=386
	RM = rm -rf
	RM_DIR = rm -rf
	MKDIR = mkdir -p
	COPY = cp
	COPY_DIR = cp -r
	fix_path = $(subst \,/,$(1))
endif

#定义全局参数
NAME = sk
OUT_DIR = ./out

#清理输出目录
clean:
	-$(RMDIR) $(call fix_path,$(OUT_DIR))
	$(MKDIR) $(call fix_path,$(OUT_DIR))

#复制配置
copy-config:
	$(COPY) config.template.json $(call fix_path,$(OUT_DIR)/config.json)

copy-licenses:
	$(COPY) LICENSE $(call fix_path,$(OUT_DIR))
	$(COPY) THIRD_PARTY_NOTICES.md $(call fix_path,$(OUT_DIR))
	$(COPY_DIR) licenses $(call fix_path,$(OUT_DIR)/licenses)


# 编译 Linux 版本
build-linux:
	$(GOENV_LINUX) go build -ldflags="-s -w" -o $(OUT_DIR)/${NAME} main.go

# 编译 windows64 版本
build-win64:
	$(GOENV_WIN64) go build -ldflags="-s -w" -o $(OUT_DIR)/${NAME}.exe main.go

# 编译 windows32 版本
build-win32:
	$(GOENV_WIN32) go build -ldflags="-s -w" -o $(OUT_DIR)/${NAME}32.exe main.go

#初始化工程
init:
	$(COPY) config.template.json config.json

# 编译所有版本
build: clean build-linux build-win64 build-win32 copy-config copy-licenses

.PHONY: build-linux build-win64 build-win32 build clean copy-config copy-licenses