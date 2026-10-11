package demo

import "careerquest/internal/domain"

// Original exercises; no employer materials or third-party course content.
func enrich(events []domain.Event) {
	type exercise struct{ task, output, check domain.Localized }
	content := map[string]exercise{
		"architecture-lab": {
			text("Design an order service that remains usable when payment processing is unavailable.", "Спроектируйте сервис заказов, который продолжает работать при сбое платёжной системы.", "Төлем жүйесі істен шыққанда да жұмыс істейтін тапсырыс сервисін жобалаңыз."),
			text("A component diagram and a one-page decision record.", "Схема компонентов и запись архитектурного решения на одну страницу.", "Компоненттер сызбасы және бір беттік архитектуралық шешім."),
			text("Describe retries, duplicate requests, and recovery after an outage; compare two alternatives.", "Опишите повторы запросов, дубликаты и восстановление; сравните два варианта решения.", "Қайта сұрауларды, қайталанған тапсырыстарды және қалпына келуді сипаттап, екі шешімді салыстырыңыз.")},
		"api-clinic": {
			text("Design an API to reserve a meeting room without double booking.", "Спроектируйте API бронирования переговорной без двойных бронирований.", "Бөлмені екі рет брондауға жол бермейтін API жобалаңыз."),
			text("Three request/response examples and an error table.", "Три примера запросов и ответов, таблица ошибок.", "Сұрау мен жауаптың үш мысалы және қателер кестесі."),
			text("Cover validation, concurrent requests, and safe retries.", "Разберите валидацию, одновременные запросы и безопасный повтор.", "Валидацияны, қатар келген сұрауларды және қауіпсіз қайталауды қарастырыңыз.")},
		"test-workbench": {
			text("Plan tests for an import that must either save all records or save nothing.", "Подготовьте тесты импорта, который сохраняет все записи либо не сохраняет ничего.", "Барлық жазбаны бірге сақтайтын немесе ештеңе сақтамайтын импортты тестілеуді жоспарлаңыз."),
			text("A risk-based test matrix with five executable cases.", "Матрица рисков и пять исполняемых тестовых сценариев.", "Тәуекелдер матрицасы және орындалатын бес тест сценарийі."),
			text("Include invalid references, duplicate IDs, rollback, and a successful round trip.", "Проверьте неверные ссылки, повторяющиеся ID, откат и экспорт с повторным импортом.", "Қате сілтемелерді, қайталанған ID-лерді, кері қайтаруды және экспорт-импорт циклін тексеріңіз.")},
		"research-sprint": {
			text("Prepare an interview about how people choose their next learning activity.", "Подготовьте интервью о том, как люди выбирают следующую учебную активность.", "Адамдардың келесі оқу әрекетін қалай таңдайтыны туралы сұхбат дайындаңыз."),
			text("A six-question guide and a fictional pilot transcript.", "Гайд из шести вопросов и вымышленный пробный диалог.", "Алты сұрақтан тұратын нұсқаулық және ойдан шығарылған сынақ сұхбаты."),
			text("Avoid leading questions; separate observations from assumptions.", "Избегайте наводящих вопросов; отделите наблюдения от предположений.", "Жетелеуші сұрақтардан аулақ болып, бақылау мен болжамды ажыратыңыз.")},
		"design-critique": {
			text("Review a fictional booking form for clarity and keyboard access.", "Проведите ревью вымышленной формы бронирования на понятность и доступность с клавиатуры.", "Ойдан шығарылған брондау пішінінің түсініктілігін және пернетақтамен қолжетімділігін тексеріңіз."),
			text("An annotated sketch with three prioritized improvements.", "Эскиз с комментариями и тремя приоритетными улучшениями.", "Түсіндірмелері бар эскиз және басым үш жақсарту."),
			text("Explain the user problem behind each change and check focus order.", "Объясните проблему пользователя для каждого изменения и проверьте порядок фокуса.", "Әр өзгерістің пайдаланушыға пайдасын түсіндіріп, фокус ретін тексеріңіз.")},
		"analysis-lab": {
			text("Interpret a fictional experiment with 40/200 conversions in A and 46/200 in B.", "Разберите вымышленный эксперимент: 40 из 200 конверсий в A и 46 из 200 в B.", "Ойдан шығарылған экспериментті талдаңыз: A тобында 200-ден 40, B тобында 200-ден 46 конверсия."),
			text("A short analysis with uncertainty and a next-step recommendation.", "Краткий анализ с неопределённостью и рекомендацией следующего шага.", "Белгісіздік пен келесі қадам ұсынысы бар қысқа талдау."),
			text("Calculate rates and absolute difference; do not claim significance from the difference alone.", "Вычислите доли и абсолютную разницу; не объявляйте значимость только по разнице.", "Үлестерді және абсолют айырманы есептеңіз; айырмаға ғана сүйеніп маңыздылықты жарияламаңыз.")},
		"query-lab": {
			text("Create fictional orders, then investigate a query filtered by customer and creation date.", "Создайте вымышленные заказы и исследуйте запрос с фильтром по клиенту и дате.", "Ойдан шығарылған тапсырыстар жасап, клиент пен күн бойынша сұрауды зерттеңіз."),
			text("SQL, EXPLAIN output, and before/after timing with an index.", "SQL, план EXPLAIN и замеры до и после добавления индекса.", "SQL, EXPLAIN жоспары және индекс қосылғанға дейінгі және кейінгі өлшемдер."),
			text("Record data size and repetitions; explain index ordering and write costs.", "Укажите размер данных и число повторов; объясните порядок полей индекса и цену записи.", "Деректер көлемі мен қайталау санын көрсетіп, индекс өрістерінің ретін және жазу құнын түсіндіріңіз.")},
		"story-circle": {
			text("Present a technical decision to a colleague unfamiliar with the project.", "Объясните техническое решение коллеге, который не знаком с проектом.", "Жобамен таныс емес әріптеске техникалық шешімді түсіндіріңіз."),
			text("A five-minute outline and a short response to feedback.", "План выступления на пять минут и короткий ответ на обратную связь.", "Бес минуттық баяндама жоспары және пікірге қысқа жауап."),
			text("State the problem, alternatives, decision, and one unresolved risk.", "Назовите проблему, варианты, решение и один нерешённый риск.", "Мәселені, нұсқаларды, шешімді және бір шешілмеген тәуекелді атаңыз.")},
		"planning-room": {
			text("Plan a two-week delivery of a fictional booking feature.", "Спланируйте двухнедельную разработку вымышленной функции бронирования.", "Ойдан шығарылған брондау функциясын екі аптада әзірлеуді жоспарлаңыз."),
			text("A small backlog, dependency map, and acceptance checklist.", "Небольшой бэклог, карта зависимостей и критерии приёмки.", "Шағын бэклог, тәуелділіктер картасы және қабылдау талаптары."),
			text("Identify the smallest useful release and one fallback if a dependency slips.", "Выделите минимальный полезный релиз и запасной вариант при задержке зависимости.", "Ең шағын пайдалы шығарылымды және тәуелділік кешіккендегі баламаны анықтаңыз.")},
		"mentor-studio": {
			text("Practice helping a fictional junior colleague investigate a failing test.", "Потренируйтесь помогать вымышленному начинающему коллеге разобраться с падающим тестом.", "Ойдан шығарылған жас әріптеске өтпейтін тестті зерттеуге көмектесуді жаттықтырыңыз."),
			text("A conversation outline and a follow-up learning goal.", "План разговора и учебная цель для следующей встречи.", "Әңгіме жоспары және келесі кездесуге арналған оқу мақсаты."),
			text("Use open questions and let the colleague choose the next debugging step.", "Используйте открытые вопросы и дайте коллеге выбрать следующий шаг отладки.", "Ашық сұрақтарды қолданып, келесі жөндеу қадамын әріптестің өзі таңдауына мүмкіндік беріңіз.")},
		"architecture-deep-dive": {
			text("Compare synchronous calls and queued events for a fictional order workflow.", "Сравните синхронные вызовы и очередь событий в вымышленном процессе заказа.", "Ойдан шығарылған тапсырыс үдерісіндегі синхронды шақырулар мен оқиғалар кезегін салыстырыңыз."),
			text("Two sequence diagrams and a tradeoff table.", "Две диаграммы последовательности и таблица компромиссов.", "Екі реттілік диаграммасы және ымыралар кестесі."),
			text("Cover consistency, ordering, retries, observability, and operational complexity.", "Разберите согласованность, порядок, повторы, диагностику и сложность эксплуатации.", "Келісімділікті, реттілікті, қайталауларды, диагностика мен пайдалану күрделілігін талдаңыз.")},
		"orientation": {
			text("Explore the fictional workspace and locate its data export and help.", "Изучите демонстрационное рабочее пространство, найдите экспорт и справку.", "Демонстрациялық жұмыс ортасын зерттеп, экспорт пен анықтаманы табыңыз."),
			text("A personal checklist of the main navigation areas.", "Личный список основных разделов навигации.", "Негізгі навигация бөлімдерінің жеке тізімі."),
			text("Distinguish an assessed skill from a calculated estimate.", "Объясните отличие оценённого навыка от расчётного.", "Бағаланған дағды мен есептік деңгейді ажыратыңыз.")},
	}
	for i := range events {
		c, ok := content[events[i].ID]
		if !ok {
			continue
		}
		events[i].Description = c.task["en"]
		events[i].Descriptions = c.task
		events[i].Deliverable = c.output
		events[i].Criteria = []domain.Localized{c.check}
	}
}
