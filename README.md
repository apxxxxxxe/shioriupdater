## shioriupdater

### これは何？

複数ゴーストの栞を一括で最新版に更新します。  
ファイル名を以下のもの以外に変更している場合は反応しません。

### 対応している栞
以下の栞とファイルが更新対象です。
- [里々](https://github.com/ponapalt/satoriya-shiori)
  - satori.dll
  - ssu.dll
  - satorite.exe
- [YAYA](https://github.com/ponapalt/yaya-shiori)
  - yaya.dll

### 使用例

1. shioriupdater.exe を図の位置に配置します。

```
.
├── shioriupdater.exe
│
├── myghost1
│   ├── install.txt
│   ├── readme.txt
│   ├── ghost
│   │   └── master
│   │       ├── descript.txt
│   │       ├── yaya.dll
│   │       └── ...
│   └── shell
│       └── master
│           ├── descript.txt
│           └── ...
│
└── myghost2
    ├── install.txt
    ├── readme.txt
    ├── ghost
    │   └── master
    │       ├── descript.txt
    │       ├── satori.dll
    │       └── ...
    └── shell
        └── master
            ├── descript.txt
            └── ...
```

2. 実行します。
3. 実行フォルダ以下の栞ファイルが更新されます。

### 注意
- 更新対象のゴースト起動中は上手くファイルが更新されません。終了してから実行してください
