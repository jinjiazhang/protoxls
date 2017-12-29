#include <iostream>
using namespace std;

#include "scheme.pb.h"
#include "SchemeStore.h"

int main()
{
    SchemeStore1<RoleScheme, int> roleStore;
    roleStore.LoadBytes("role_data.bytes");
    const RoleScheme* config = roleStore.GetConfig(10001);
    if (config == NULL) {
        std::cout << "config not found" << std::endl;
        return -1;
    }

    std::cout << "roleId = " << config->roleid() << std::endl;
    std::cout << "attack = " << config->props().attack() << std::endl;
    std::cout << "defense = " << config->props().defense() << std::endl;
    return 0;
}