[![GPL 3.0 license](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://github.com/REIJI007/AdBlock_Rule_For_V2ray/blob/main/LICENSE-GPL%203.0)
[![CC BY-NC-SA 4.0 license](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg)](https://github.com/REIJI007/AdBlock_Rule_For_V2ray/blob/main/LICENSE-CC-BY-NC-SA%204.0)
<!-- 居中的大标题 -->
<h1 align="center" style="font-size: 70px; margin-bottom: 20px;">AdBlock_Rule_For_V2ray</h1>

<!-- 居中的副标题 -->
<h2 align="center" style="font-size: 30px; margin-bottom: 40px;">适用于V2ray（V2ray核心与Xray核心）的广告域名拦截adblock.dat二进制文件，每20分钟更新一次</h2>

<!-- 徽章（根据需要调整） -->
<p align="center" style="margin-bottom: 40px;">
    <img src="https://img.shields.io/badge/last%20commit-today-brightgreen" alt="last commit" style="margin-right: 10px;">
    <img src="https://img.shields.io/github/forks/REIJI007/AdBlock_Rule_For_V2ray" alt="forks" style="margin-right: 10px;">
    <img src="https://img.shields.io/github/stars/REIJI007/AdBlock_Rule_For_V2ray" alt="stars" style="margin-right: 10px;">
    <img src="https://img.shields.io/github/issues/REIJI007/AdBlock_Rule_For_V2ray" alt="issues" style="margin-right: 10px;">
    <img src="https://img.shields.io/github/license/REIJI007/AdBlock_Rule_For_V2ray" alt="license" style="margin-right: 10px;">
</p>

**一、从多个广告过滤器中提取拦截域名条目，删除重复项，并将它们转换为兼容V2ray的dat二进制文件，其中列表的每个条目都写成了形如domain:example.com形式，一行仅一条规则。该列表可以用作V2ray的拦截域名路由文件，以阻止广告域名， powershell脚本和main.go转换程序每20分钟自动执行，并将生成的文件发布在release中.两个文件的下载地址分别如下，其中adblock_reject_domain.txt由powershell脚本生成，adblock.dat则是由main.go转换程序将adblock.txt转化得来的dat二进制文件，该文件仅有一个域名标签```ADBLOCK```**
<br>
<br>

<table border="1" style="border-collapse: collapse; width: 100%; font-family: Arial, sans-serif;">
  <tr>
    <td colspan="2" style="background-color: #f2f2f2; font-weight: bold; text-align: center; padding: 10px;">订阅地址</td>
  </tr>
  <tr>
    <td style="padding: 8px;">DAT</td>
    <td style="padding: 8px;">
      <strong><a href="https://raw.githubusercontent.com/REIJI007/AdBlock_Rule_For_V2ray/main/adblock.dat" style="color: #0066cc;">Github原始链接</a></strong> | 
      <strong><a href="https://www.adblock-v2ray-dat.reiji007.org" style="color: #0066cc;">Cloudflare加速链接</a></strong>
    </td>
  </tr>
  <tr>
    <td style="padding: 8px;">拦截域名</td>
    <td style="padding: 8px;">
      <strong><a href="https://raw.githubusercontent.com/REIJI007/AdBlock_Rule_For_V2ray/main/adblock_reject_domain.txt" style="color: #0066cc;">Github原始链接</a></strong> | 
      <strong><a href="https://www.adblock-v2ray-reject-domain.reiji007.org" style="color: #0066cc;">Cloudflare加速链接</a></strong>
    </td>
  </tr>
</table>


<hr>

## 警告:本过滤器订阅有可能破坏某些网站的功能，也有可能封禁某些色情、赌博网站，使用前请斟酌考虑，如有误杀请积极向上游issue反馈，本仓库仅提供去重、筛选、合并功能

<hr>

**二、可加入此powershell脚本处理的有```adblock plus语法广告过滤器```、```Host拦截列表```、```纯广告域名列表```、```Dnsmasq列表```，请自行酌情添加过滤器订阅URL至powershell脚本中进行处理，你可将该脚本代码复制到本地文本编辑器制作成.ps1后缀的文件运行在powershell上，注意修改生成的文本文件路径，最后在V2ray的json配置中加入被拦截域名，且V2ray配置字段写成类似于如下例子**
<br>
<br>
*简而言之就是可以让你DIY出希望得到的拦截域名列表，缺点是此做法只适合本地定制使用，当然你也可以像本仓库一样部署到GitHub上面，见仁见智*
<br>
<br>

**三、本仓库引用多个广告过滤器，从这些广告过滤器中提取了被拦截条目的域名，剔除了非拦截项并去重，最后做成adblock.dat二进制文件，虽无法做到面面俱到但能减少广告带来的困扰，请自行斟酌考虑使用。碍于V2ray的路由行为且秉持着尽可能不误杀的原则，本仓库采取域名后缀匹配策略，即匹配命中于拦截列表上的域名或其子域名时触发拦截，除此之外的情况给予放行，尽管这会有许多漏网之鱼的广告被放行**
<br>
<br>

**四、关于本仓库使用方式：**

  *使用方式一：下载releases中的adblock_reject_domain.txt文件，修改V2ray的json配置中的"routing"字段下的"domain"部分*
  
<hr>


```conf

{
 "outbounds": 
   [
     {
       "protocol": "blackhole",
       "tag": "adblock"          // 此outboundTag出站配合下面的域名拦截路由
     }
   ],
 "routing": 
   {
     "domainStrategy": "AsIs",
     "rules": 
     [
       {
         "type": "field",
         "domain": 
         [
             "example.com1",   // 在这里替换要拦截出站的广告域名,注意最后一个广告条目不用加逗号
             "example.com2",
             "example.com3"
         ],
         "outboundTag": "adblock"  // 匹配到的域名流量会被导流到名为adblock的outboundTag出站
       }
     ]
   }
}
```
<hr>

  *使用方式二：下载adblock.dat文件到V2ray同目录下，将下面对应格式的配置文件中"outbounds"字段和"routing"字段内容添加到你的json配置文件中，注意"outbounds"与"routing"之间的配合，注意去掉注释，"tag" 值需要保持一致*
<hr>


```conf
{
    "outbounds": 
    [
        {
            "protocol": "blackhole",
            "tag": "adblock"  // 此 outboundTag 出站配合下面的域名拦截路由
        }
    ],
    "routing": 
    {
        "domainStrategy": "AsIs",
        "rules": 
        [
            {
                "type": "field",
                "domain": 
                [
                    "ext:adblock.dat:adblock"  // 引用 adblock.dat 文件中的 adblock 标签
                ],
                "outboundTag": "adblock"  // 匹配到的域名流量会被导流到名为 adblock 的 outboundTag 出站
            }
        ]
    }
}
```
<hr>

**五、关于本仓库的使用效果为什么没有普通广告过滤器效果好的疑问解答：**
<br>
*因为普通的广告过滤器包含域名过滤（拦截广告域名）、路径过滤（例如拦截URL路径中包含/ads/的所有请求）、正则表达式过滤（例如拦截所有包含ads.js或ad.js的URL）、类型过滤（例如只拦截图片资源）、隐藏元素等等多因素作用下使得在广告拦截测试网站中可以取得高分。**但碍于V2ray的路由行为（可参考相关文档）**，本仓库仅提取了被拦截域名进行域名匹配过滤，换言之，本仓库就是一个“删减版”的广告过滤器（仅保留了域名匹配过滤功能，规则数在**15万**条左右），所以最终效果没有广告过滤器效果好*
<br>
<br>



**六、本仓库引用的广告过滤规则来源请查看```Referencing rule sources.txt```（目前107个来源）。至于是否误杀域名完全取决于这些处于上游的广告过滤器的域名拦截行为，若不满意的话可按照第二条在本地使用powershell脚本进行DIY本地定制化拦截域名列表，亦或可以像本仓库一样DIY定制后部署到github上面，或者fork本仓库自行DIY**


**七、特别鸣谢**



1. [v2ray](https://github.com/v2fly/v2ray-core)
2. [Adguard](https://github.com/AdguardTeam/AdGuardFilters)





## LICENSE
- [CC-BY-SA-4.0 License](https://github.com/REIJI007/AdBlock_Rule_For_V2ray/blob/main/LICENSE-CC-BY-NC-SA%204.0)
- [GPL-3.0 License](https://github.com/REIJI007/AdBlock_Rule_For_V2ray/blob/main/LICENSE-GPL%203.0)
