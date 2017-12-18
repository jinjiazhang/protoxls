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
    setlocale(LC_ALL, "chs");

    if (argc < 2) {
        help();
        return -1;
    }

    ProtoExcel runner;
    if (!runner.ParseScheme(argv[1])) {
        std::cerr << "protoxls parse proto fail" << std::endl;
        return -1;
    }

    if (!runner.ExportResult()) {
        std::cerr << "protoxls exprot result fail" << std::endl;
        return -3;
    }
    return 0;
}