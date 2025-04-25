# Bánh

Ngôn ngữ lập trình bằng tiếng Việt

Thiết kế cuối cùng ↓

```banh
biến a, b E N32

hàm chính() -> N32
  trong khi a <= b hoặc b >= a và a = 1 thì
    in("nhỏ hơn")
    a := a + 1; in(b)
  kết thúc
kết thúc

hàm cộng(a E N32, b E N32) -> N32
  trả về a + b
kết thúc
```

Hiện trạng: Cài đặt backend LLVM

- Tính năng:
  [x] Tạo biến
  [x] trả giá trị
  [] Sử dụng hàm tạo ra
  [] In giá trị ra
  [] v.v...
