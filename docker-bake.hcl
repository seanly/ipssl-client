variable "VERSION" {
  default = "0.0.1"
}

variable "FIXID" {
  default = "1"
}

group "default" {
  targets = ["ipssl-client"]
}

target "ipssl-client" {
    labels = {
        "cloud.opsbox.author" = "seanly"
        "cloud.opsbox.image.name" = "ipssl-client"
        "cloud.opsbox.image.version" = "${VERSION}"
        "cloud.opsbox.image.fixid" = "${FIXID}"
    }
    dockerfile = "Dockerfile"
    context  = "./"
    args = {
        ipssl-client_VERSION="${VERSION}"
    }
    platforms = ["linux/amd64", "linux/arm64"]
    tags = [
        "seanly/appset:ipssl-client-${VERSION}-${FIXID}",
        "seanly/appset:ipssl-client-${VERSION}"
    ]
    output = ["type=image,push=true"]
}

variable "ACR_REGISTRY" {
  default = "registry.cn-chengdu.aliyuncs.com"
}

group "acr" {
  targets = ["ipssl-client-amd64", "ipssl-client-arm64"]
}

target "ipssl-client-amd64" {
    labels = {
        "cloud.opsbox.author" = "seanly"
        "cloud.opsbox.image.name" = "ipssl-client"
        "cloud.opsbox.image.version" = "${VERSION}"
        "cloud.opsbox.image.fixid" = "${FIXID}"
    }
    dockerfile = "Dockerfile"
    context  = "./"
    args = {
        ipssl-client_VERSION="${VERSION}"
    }
    platforms = ["linux/amd64"]
    tags = [
        "${ACR_REGISTRY}/seanly/appset:ipssl-client-${VERSION}-${FIXID}",
        "${ACR_REGISTRY}/seanly/appset:ipssl-client-${VERSION}"
    ]
    output = ["type=image,push=true"]
}

target "ipssl-client-arm64" {
    labels = {
        "cloud.opsbox.author" = "seanly"
        "cloud.opsbox.image.name" = "ipssl-client"
        "cloud.opsbox.image.version" = "${VERSION}"
        "cloud.opsbox.image.fixid" = "${FIXID}"
    }
    dockerfile = "Dockerfile"
    context  = "./"
    args = {
        ipssl-client_VERSION="${VERSION}"
    }
    platforms = ["linux/arm64"]
    tags = [
        "${ACR_REGISTRY}/seanly/appset:ipssl-client-${VERSION}-${FIXID}-arm64",
        "${ACR_REGISTRY}/seanly/appset:ipssl-client-${VERSION}-arm64"
    ]
    output = ["type=image,push=true"]
}
