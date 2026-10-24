package mirror

// Provider 下载源
type Provider struct {
	UseMirror    bool
}

const (
	mojangBase   = "https://piston-meta.mojang.com"
	mojangAssets = "https://resources.download.minecraft.net"
	mojangLib    = "https://libraries.minecraft.net"
	bmclBase     = "https://bmclapi2.bangbang93.com"
)

// VersionManifestURL 版本清单
func (p Provider) VersionManifestURL() string {
	if p.UseMirror {
		return bmclBase + "/mc/game/version_manifest_v2.json"
	}
	return mojangBase + "/mc/game/version_manifest_v2.json"
}

// VersionJSONURL 单版本详情
func (p Provider) VersionJSONURL(raw string) string {
	if p.UseMirror {
		// BMCL 镜像：直接用版本号拼
		return bmclBase + "/mc/" + raw
	}
	return raw
}

// AssetURL 资源文件
func (p Provider) AssetURL(hash string) string {
	base := mojangAssets
	if p.UseMirror {
		base = bmclBase
	}
	return base + "/" + hash[:2] + "/" + hash
}

// LibraryURL 库文件
func (p Provider) LibraryURL(raw string) string {
	if p.UseMirror && !isURL(raw) {
		return bmclBase + "/maven/" + raw
	}
	if isURL(raw) {
		return raw
	}
	return mojangLib + "/" + raw
}

// FabricMetaURL Fabric 元数据
func (p Provider) FabricMetaURL() string {
	if p.UseMirror {
		return bmclBase + "/fabric-meta/v2/versions/loader"
	}
	return "https://meta.fabricmc.net/v2/versions/loader"
}

// ModrinthURL Modrinth API
func (p Provider) ModrinthURL() string {
	return "https://api.modrinth.com/v2"
}

func isURL(s string) bool {
	return len(s) > 8 && (s[:8] == "https://" || s[:7] == "http://")
}
