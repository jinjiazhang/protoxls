#ifndef _JINJIAZHANG_STRCONV_H_
#define _JINJIAZHANG_STRCONV_H_

#include <Windows.h>
#include <string>
#include <tchar.h>

#define ansi2unicode(s)			(m2wstring((s), CP_ACP))
#define unicode2ansi(s)			(w2mstring((s), CP_ACP))
#define utf82unicode(s)			(m2wstring((s), CP_UTF8))
#define unicode2utf8(s)			(w2mstring((s), CP_UTF8))
#define ansi2utf8(s)			(unicode2utf8(ansi2unicode((s))))
#define utf82ansi(s)			(unicode2ansi(utf82unicode((s))))

std::wstring m2wstring(std::string src, int codepage);
std::string w2mstring(std::wstring src, int codepage);

#endif