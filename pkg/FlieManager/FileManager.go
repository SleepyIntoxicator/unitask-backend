package FlieManager

import "os"

//FileManager - manage user files for tasks
//Менеджер добавленных пользовательских материалов
//tbl Файл - путь к файлу, название, indexHash, markAsDeleted
//tbl Ссылка на файл - университет, группа, хэш-файла, добавил_id
type FileManager struct {
	basePath string
}

type IFileManager interface {
	AddFile(fileName string, content os.File)
	MarkAsDeleted(fileHash string, userID, groupID int)

	RenameFile(fileHash string, newFileName string)

	FindFile(fileName string, groupID int)
	FindFile(fileName string, userID int)

	AssociateToUser(fileHash string, userID int)
	AssociateToGroup(fileHash string, groupID int)

}

func NewFileManager(basePath string) *FileManager {
	return &FileManager{
		basePath: basePath,
	}
}

func (mgr *FileManager) F() {

}