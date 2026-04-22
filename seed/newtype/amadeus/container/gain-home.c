#include <errno.h>
#include <pwd.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

int main() {
  struct passwd *pw = getpwnam("amadeus");
  if (pw == NULL) {
    return 1;
  }
  if (mkdir("/home/amadeus", 0750) != 0 && errno != EEXIST) {
    return 1;
  }
  if (chown("/home/amadeus", pw->pw_uid, pw->pw_gid) != 0) {
    return 1;
  }
  if (chmod("/home/amadeus", 0750) != 0) {
    return 1;
  }
  return 0;
}
