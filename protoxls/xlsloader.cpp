#include <stdio.h>
#include <stdarg.h>

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