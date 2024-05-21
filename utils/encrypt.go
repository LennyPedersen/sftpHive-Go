package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "flag"
    "fmt"
    "io"
    "os"
)

func encrypt(plainText, key string) (string, error) {
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    cipherText := make([]byte, aes.BlockSize+len(plainText))
    iv := cipherText[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(plainText))

    return base64.URLEncoding.EncodeToString(cipherText), nil
}

func main() {
    key := "aReaLlYHaxxIEncryptionKey" // Must be 32 bytes long for AES-256, please change this to a secure key :)

    password := flag.String("password", "", "The customer's SFTP password to encrypt")
    flag.Parse()

    if *password == "" {
        fmt.Println("Please provide a password using the -password flag")
        os.Exit(1)
    }

    encryptedPassword, err := encrypt(*password, key)
    if err != nil {
        fmt.Println("Error encrypting password:", err)
        return
    }

    fmt.Println("Encrypted password:", encryptedPassword)
}
