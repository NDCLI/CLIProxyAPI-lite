# Kế hoạch cải tiến giao diện quản trị

Mục tiêu: người dùng nhận ra tác vụ qua icon và nhãn, đọc trạng thái tài khoản ngay trên thẻ, và thấy phản hồi rõ sau mỗi thao tác. Giao diện giữ sự gọn gàng của một bảng điều khiển vận hành; hiệu ứng chỉ xuất hiện khi người dùng tương tác hoặc dữ liệu thay đổi.

Đã làm trong đợt này: icon cho các nút chính, hiệu ứng hover/nhấn/focus có hỗ trợ giảm chuyển động, tự nhận callback OAuth cho giao diện WebUI, ô dán URL callback Codex khi cần, và trạng thái lỗi/đăng nhập lại trên thẻ quota GPT. Phần còn lại triển khai theo thứ tự bên dưới.

## 1. Nền tảng chung

- Chuẩn hóa nút chính, phụ, nguy hiểm và nút chỉ có icon. Mọi nút chỉ có icon phải có tên truy cập được và tooltip.
- Dùng bộ SVG hiện có cho các tác vụ nhất quán: thêm, sao chép, lưu, kiểm tra, làm mới, bật/tắt, sửa và xóa.
- Thêm trạng thái hover, nhấn, focus bàn phím, đang xử lý, thành công và lỗi. Hiệu ứng ngắn, không làm dịch chuyển bố cục, tôn trọng `prefers-reduced-motion`.
- Giữ màu trạng thái nhất quán: xanh cho thành công/còn hạn mức, vàng cho cần chú ý, đỏ cho lỗi. Màu không được là tín hiệu duy nhất.

## 2. Theo dõi hạn mức và tài khoản

- Hiển thị tài khoản thành thẻ gọn theo kiểu Cockpit, với tên, nhà cung cấp, trạng thái và các nhóm hạn mức dễ đọc.
- Lấy model Antigravity đang hiện trong registry/IDE; bỏ model ảnh và model nội bộ. Chỉ dùng dữ liệu quota thực từ upstream, không suy ra cửa sổ 5 giờ/tuần từ thời điểm reset.
- Codex/GPT: tự đọc quota khi mở tab. Nếu OAuth hết hiệu lực, hiện trạng thái cụ thể và đường đăng nhập lại ngay trên thẻ; sau khi kết nối lại, làm mới tự động.
- Nút làm mới toàn trang và trên từng tài khoản giữ vị trí cố định. Tài khoản đang tải không làm các thẻ khác chờ.

## 3. Từng khu vực

- Tổng quan và Sử dụng: ưu tiên số liệu thực, xu hướng và lỗi cần xử lý; giảm các khung chỉ mang tính trang trí.
- Endpoint & Khóa: phân biệt rõ thao tác tạo/sao chép/xoay/xóa; phản hồi sao chép và cảnh báo trước thao tác nguy hiểm.
- Nhà cung cấp và File xác thực: logo, tìm kiếm, trạng thái, model và hành động cùng một cấu trúc hàng; thông tin dài được rút gọn có tooltip.
- Combo & Vision: đường fallback đọc theo thứ tự, trạng thái kiểm tra rõ, sửa và nhân bản không làm mất biểu mẫu.
- Log, Cài đặt, Plugin và Công cụ CLI: bộ lọc, trạng thái rỗng/lỗi, nút áp dụng/lưu và phản hồi tác vụ dùng cùng mẫu điều khiển.
- Bắt đầu nhanh và Tiết kiệm token: hướng dẫn ngắn theo bước có thể thực hiện, nêu rõ tính năng nào chưa có backend.

## 4. Kiểm tra trước khi giao

- Kiểm tra bằng chuột và bàn phím ở màn hình rộng, laptop và điện thoại; không có nút mất tên hoặc bị che.
- Kiểm tra dữ liệu đang tải, rỗng, lỗi và thành công ở các trang chính; không hiển thị số liệu giả.
- Chạy kiểm tra JavaScript, test Go liên quan và build server. Đối chiếu ảnh chụp giao diện trước/sau ở các trang có sửa.
