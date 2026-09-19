package data

// IndexFile 数据根目录下的索引文件名。各业务域的数据文件由索引显式列出，
// 加载器不再扫描目录，因此哪些数据生效由数据自己决定。
const IndexFile = "metadata.json"

// Index 数据根目录的索引：各业务域的数据文件路径一律相对数据根。
type Index struct {
	Bid  BidFiles  `json:"bid"`
	Pool PoolFiles `json:"pool"`
}

// BidFiles 竞拍清单域索引的文件列表。
type BidFiles struct {
	List []string `json:"list"`
}

// PoolFiles 盲盒池域索引的文件列表，以及池未自带资源时回退的全局资源文件。
type PoolFiles struct {
	Global GlobalFiles `json:"global"`
	List   []string    `json:"list"`
}

// GlobalFiles 索引里域内共享的数据文件。
type GlobalFiles struct {
	Resources string `json:"resources"`
}

// LoadIndex 从来源读取索引文件。
func LoadIndex(source Source) (Index, error) {
	var index Index

	if err := ReadJSON(source, IndexFile, &index); err != nil {
		return Index{}, err
	}

	return index, nil
}
