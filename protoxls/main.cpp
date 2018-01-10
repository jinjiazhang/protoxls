#include <stdio.h>
#include <stdarg.h>
#include <iostream>
#include "ProtoExcel.h"
using namespace std;

void help()
{

}

void proto_log(int level, const char* format, ...)
{
    va_list	args;
    va_start(args, format);
#ifdef _WIN32
    _vprintf_p(format, args);
#else
    vprintf(format, args);
#endif // _WIN32
    va_end(args);
}

// ./protoxls scheme.proto
int main(int argc, char* argv[])
{
    if (argc < 2) {
        help();
        return -1;
    }

    ProtoExcel runner;
	char* proto = argv[1];
    if (!runner.ParseScheme(proto)) {
        proto_error("protoxls parse scheme fail, proto=%s\n", proto);
        return -1;
    }

    if (!runner.ExportResult()) {
        proto_error("protoxls exprot result fail, proto=%s\n", proto);
        return -2;
    }
    return 0;
}
