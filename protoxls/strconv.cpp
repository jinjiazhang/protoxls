#include "strconv.h"

std::wstring m2wstring(std::string src, int codepage)
{
    int size = ::lstrlenA(src.c_str()) + 1;
    int length = ::MultiByteToWideChar(codepage, 0, src.c_str(), size, NULL, 0);
    wchar_t* temp = new wchar_t[length + 1];
    memset(temp, 0, (length + 1) * sizeof(wchar_t));
    ::MultiByteToWideChar(codepage, 0, src.c_str(), size, (LPWSTR)temp, length);
    std::wstring dest(temp);
    delete [] temp;
    return dest;
}

std::string w2mstring(std::wstring src, int codepage)
{
    int size = ::lstrlenW(src.c_str()) + 1;
    int length = ::WideCharToMultiByte(codepage, 0, src.c_str(), size, NULL, 0, NULL, NULL);
    char* temp = new char[length + 1];
    memset((void*)temp, 0, (length + 1) * sizeof(char));
    ::WideCharToMultiByte(codepage, 0, src.c_str(), size, (LPSTR)temp, length, NULL, NULL);
    std::string dest(temp);
    delete [] temp;
    return dest;
}