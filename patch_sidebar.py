with open("/home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin/src/lib/components/Sidebar.svelte", "r") as f:
    content = f.read()

import re
# Remove all dark: classes
content = re.sub(r'\s*dark:[\w\-\/]+', '', content)
# Simplify active state slightly to look more like CBT
# and remove custom-scrollbar from style? No, just keep dark: removal.

with open("/home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin/src/lib/components/Sidebar.svelte", "w") as f:
    f.write(content)
