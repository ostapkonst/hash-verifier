package checksum

type VerifyStatusType int

const (
	HashMatched VerifyStatusType = iota
	HashMismatch
	Unreadable
)

func (v VerifyStatusType) String() string {
	switch v {
	case HashMatched:
		return "MATCHED"
	case HashMismatch:
		return "MISMATCH"
	case Unreadable:
		return "UNREADABLE"
	default:
		panic("unknown status")
	}
}

// VerifyResult — результат проверки одного файла.
type VerifyResult struct {
	Path         string           // относительный путь
	ActualHash   string           // вычисленный хеш
	ExpectedHash string           // ожидаемый хеш
	Status       VerifyStatusType // статус сравнения хешей
	ReadBytes    int64            // количество прочитанных байт файла при вычислении хеша
	Err          error            // ошибка при вычислении хеша
}

// GenerateResult — результат генерации хеша для одного файла.
type GenerateResult struct {
	RelPath   string // относительный путь с префиксом или без него
	Hash      string // вычисленный хеш
	ReadBytes int64  // количество прочитанных байт файла при вычислении хеша
	Err       error  // ошибка при вычислении хеша
}

// Статистика для генератора.
type GeneratorStats struct {
	TotalFiles          int     // всего файлов в чек-сумме
	Processed           int     // обработано успешно
	WithErrors          int     // не удалось обработать
	CurrentFileOrStatus string  // текущий файл или статус
	FileHashingProgress float64 // прогресс вычисления хеша текущего файла
}

func (g GeneratorStats) Pending() int { return g.TotalFiles - g.Processed - g.WithErrors }

func (g GeneratorStats) TotalProgress() float64 {
	if g.TotalFiles == 0 {
		return 0
	}

	return float64(g.TotalFiles-g.Pending()) / float64(g.TotalFiles)
}

// Статистика для верификатора.
type VerifierStats struct {
	TotalFiles          int     // всего файлов в чек-сумме
	Matched             int     // проверено успешно
	Mismatch            int     // не прошло проверку
	Unreadable          int     // не удалось проверить
	CurrentFileOrStatus string  // текущий файл или статус
	FileHashingProgress float64 // прогресс вычисления хеша текущего файла
}

func (v VerifierStats) Pending() int { return v.TotalFiles - v.Matched - v.Mismatch - v.Unreadable }

func (v VerifierStats) TotalProgress() float64 {
	if v.TotalFiles == 0 {
		return 0
	}

	return float64(v.TotalFiles-v.Pending()) / float64(v.TotalFiles)
}
