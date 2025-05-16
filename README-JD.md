# 说明
当前集群社区release-0.11分支开发迭代,release-0.11即为主分支
# 发布版本
## 版本号
以**v0.11-jd-0.1.0**为例  
- v0.11：对应社区的release-0.11  
- jd-0.1.0：jd后的三位分别是“主版本号”、“次版本号”、“修订版本号”
## 0.运行测试套件确保所有测试通过
```shell
make test
```
## 1.创建发布分支
```shell
# 设置必要的环境变量
export NEW_VERSION=xxx # 如，v0.11-jd-0.1.0
export GIT_TAG=${NEW_VERSION}
export IMAGE_REGISTRY=hub.jdcloud.com/jdos  # 使用你的镜像仓库
export IMAGE_REPO=${IMAGE_REGISTRY}/kueue

# 从主分支创建发布分支
git checkout -b release-${NEW_VERSION}

# 推送发布分支到远程仓库
git push origin release-${NEW_VERSION}
```
## 2.构建发布制品
```shell
# 构建发布制品
make artifacts
```
这将生成：
- 各种配置的 Kubernetes manifests 文件
- 打包的 Helm Chart
- 多平台的 CLI 工具
## 3.构建并推送 Docker 镜像
```shell
# 构建 Docker 镜像
make image-build

# 为镜像添加版本标签
docker tag ${IMAGE_REPO}:latest ${IMAGE_REPO}:${NEW_VERSION}

# 推送到容器仓库
docker push ${IMAGE_REPO}:${NEW_VERSION}
docker push ${IMAGE_REPO}:latest
```
## 4.创建 Git 标签和发布
```shell
# 创建带注释的 Git 标签
git tag -a ${NEW_VERSION} -m "Kueue ${NEW_VERSION}"

# 推送标签到远程仓库
git push origin ${NEW_VERSION}
```
然后，在 GitHub/GitLab 界面上：
创建新的发布（Release）
选择刚创建的标签
填写发布说明，包含版本亮点和 CHANGELOG 中的信息
上传构建好的制品（artifacts 目录中的文件）
## 5.后续维护
```shell
# 将发布分支的变更合并回主分支(如果有)
git checkout release-0.11
git merge release-${NEW_VERSION}
git push origin release-0.11

# 准备下一个开发版本
# 例如更新版本常量为下一个开发版本
```
