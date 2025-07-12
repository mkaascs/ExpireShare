# **ExpireShare** 🔥📁
### *A self-destructing file-sharing service with TTL and download limits.*

**ExpireShare** lets you upload files and share them via short-lived links. Set expiration time (`TTL`) or max downloads, and the file auto-deletes when limits are hit—no traces left!

## **Features**
- ⏳ **Time-to-Live (TTL)**: Files vanish after `N` hours/days.
- 🔢 **Download limits**: Expire after `X` downloads.
- 🔐 **Optional passwords**: Protect files with hashed passwords.
- 🤖 **Telegram Bot**: Upload/files management via chat.
- ⌨️ **CLI Tool**: Manage shares from the terminal.
- 📊 **REST API**: Integrate with your apps.

## **Tech Stack**
- **Backend**: Golang + Chi
- **Storage**: Local disk (temporarily)
- **DB**: MySQL
- **Auth**: Bcrypt (password hashing)

## **Use Cases**
- Send sensitive documents (*expire after 1 download*).
- Share temporary media (*delete in 24h*).
- Replace WeTransfer with custom rules.

```bash  
# CLI example  
expireshare upload secret.pdf --ttl=1h --downloads=3 --password=123  
```  

---  

### **Why ExpireShare?**
Unlike permanent cloud storage, ExpireShare prioritizes **privacy** and **auto-cleanup**. Perfect for developers, teams, or anyone who values control over shared files.