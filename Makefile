all: build-windows

ENABLE_UNSUBSCRIBE ?= 0
WAILS_TAGS := webkit2_41

ifeq ($(ENABLE_UNSUBSCRIBE),1)
WAILS_TAGS += enable_unsubscribe
endif

build-windows:
	wails build -platform windows/amd64 -tags "$(WAILS_TAGS)"
