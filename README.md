# Bánh 🥖
> "Dễ như ăn 'Bánh'"
>
> **- Bánh, 2025**

Ngôn ngữ lập trình bằng tiếng Việt

Thiết kế cuối cùng ↓

```banh
biến a, b E Z32

hàm chính() -> Z32
  a := 5
  b := 7
  trong khi a <= b hoặc a != 10 thì
    in("nhỏ hơn")
    a := a + 1; in(cộng(a, b))
  kết thúc
kết thúc

hàm cộng(a E Z32, b E Z32) -> Z32
  trả về a + b
kết thúc
```

## Thông tin
Là Frontend LLVM, sử dụng LLC + Clang với LLI ngầm để tạo chương trình/chạy code.


## Tính năng

- [x] Tạo biến

- [x] Trả giá trị

- [x] Sử dụng hàm tạo ra

- [x] Nếu/Không thì

- [x] In giá trị ra

- [x] Mảng

- [ ] Chuỗi

- [ ] Dữ liệu có cấu trúc

- [ ] Thư viện sẵn

- [ ] Chương trình nhiều tệp nguồn

- [ ] Ma trận

- [ ] Sử dụng hàm ffi (?)

- [ ] V.V...
