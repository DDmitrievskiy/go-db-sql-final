package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	// get
	// получаем только что добавленную посылку, проверяем отсутствие ошибки
	// проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, parcel, stored)
	// delete
	// удаляем добавленную посылку, проверяем отсутствие ошибки
	// проверяем, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	require.NoError(t, err)
	_, err = store.Get(parcel.Number)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	// set address
	// обновляем адрес, проверяем отсутствие ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)
	// check
	// получаем добавленную посылку и проверяем, что адрес обновился
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")

	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	// set status
	// обновляем статус, проверяем отсутствие ошибки
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)
	// check
	// получаем добавленную посылку и проверяем, что статус обновился
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, stored.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")

	require.NoError(t, err)
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	store := NewParcelStore(db)
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавляем новую посылку в БД, проверяем отсутствие ошибки и наличие идентификатора
		require.NoError(t, err)
		require.NotEmpty(t, id)
		// обновляем идентификатор у добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получаем список посылок по идентификатору клиента, сохранённого в переменной client
	// проверяем отсутствие ошибки
	// проверяем, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// проверяем, что все посылки из storedParcels есть в parcelMap
		// проверяем, что значения полей полученных посылок заполнены верно
		require.Contains(t, parcelMap, parcel.Number)
		require.Equal(t, parcelMap[parcel.Number], parcel)
	}
}
