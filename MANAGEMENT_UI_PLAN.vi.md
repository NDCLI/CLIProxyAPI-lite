# Kế hoạch cải tiến giao diện quản trị

Mục tiêu: người dùng nhận ra tác vụ qua icon và nhãn, đọc trạng thái tài khoản ngay trên thẻ, và thấy phản hồi rõ sau mỗi thao tác. Giao diện giữ sự gọn gàng của một bảng điều khiển vận hành; hiệu ứng chỉ xuất hiện khi người dùng tương tác hoặc dữ liệu thay đổi.

Trạng thái 23/09/2026: các mục giao diện 1–3 đã triển khai. Nút và trạng thái dùng cùng một hệ màu/icon; quota tự tải, gộp theo nhóm với danh sách model có thể mở rộng, và chỉ báo rõ khi Codex cần đăng nhập lại. Tổng quan, sử dụng, endpoint, nhà cung cấp, file xác thực, combo, log, cài đặt, plugin và công cụ CLI đã có bộ lọc hoặc phản hồi tác vụ phù hợp. Tiết kiệm token nêu rõ giới hạn backend hiện tại. Kiểm tra mã nguồn, bản build và dữ liệu trực tiếp trên màn hình rộng đã hoàn tất; kiểm tra trực quan trên thiết bị di động thật vẫn cần thực hiện.

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

## 5. Đợt tối ưu tương tác — kế hoạch ngày 24/09/2026

Trạng thái: **chỉ lập kế hoạch, chưa triển khai**. Theo yêu cầu mới nhất, dừng build/test; mục này thay thế yêu cầu chạy build/test ở mục 4 cho đợt công việc hiện tại. Không khởi động server, MITM hoặc tác động DNS/chứng chỉ để khảo sát giao diện.

### 5.1. Những điểm đã thấy trong mã nguồn

- `internal/managementasset/web/tools.js`: `renderMITM()` xóa nội dung trang để hiện Loading, tải status/khóa/model rồi mới tải mappings. Đổi công cụ và nhiều thao tác lại gọi toàn bộ luồng này. Đây là nguyên nhân có thể gây nháy và chờ nhiều nhịp; chưa tái hiện trực tiếp để kết luận có lỗi modal.
- `internal/managementasset/web/app.css`: animation xuất hiện được áp vào nhiều khối; dựng lại DOM có thể làm animation chạy lại. CSS còn dùng con trỏ chờ cho mọi nút disabled, chưa phân biệt “không có thay đổi” với “đang xử lý”.
- Lưu mapping MITM, Cài đặt và Plugin khóa nút lúc gửi nhưng luôn bật lại trong `finally`, chưa dựa trên khác biệt so với dữ liệu đã lưu. Combo có cờ dirty phục vụ cảnh báo đóng, cần đồng bộ với trạng thái nút Lưu.
- Đã có `flashAction()` đổi sang dấu tích và nhãn thành công trong 2.200 ms. Tuy nhiên Copy guide/preview ở CLI Tools dùng `event.currentTarget` sau `await`; khi đó tham chiếu có thể đã thành `null`, khiến phản hồi lỗi dù clipboard đã được ghi.

### 5.2. Thứ tự triển khai

#### Bước 1 — Sửa Copy và thống nhất phản hồi nút

- Lưu tham chiếu nút trước `await`; tái sử dụng `copyText()` và `flashAction()`, không thêm thư viện.
- Chỉ hiện dấu tích + “Đã sao chép” sau khi clipboard thành công; giữ khoảng 2,2 giây rồi trở lại. Lỗi quyền clipboard phải báo “Không sao chép được”, không báo thành công giả.
- Áp dụng nhất quán cho các nút Copy hiện có ở Bắt đầu nhanh, Endpoint/Khóa và CLI Tools; bấm lại sau khi hoàn tất sẽ tính lại thời gian phản hồi.
- Giữ độ rộng nút, focus và bố cục khi đổi nhãn; phân biệt disabled với busy, bổ sung trạng thái đọc được bằng bàn phím/trình đọc màn hình. Giữ hỗ trợ giảm chuyển động và bản dịch Việt/Anh; không đưa nội dung khóa vào thông báo.

#### Bước 2 — Chuẩn hóa Lưu theo dữ liệu thay đổi

| Trạng thái | Cách hiển thị và hành vi |
| --- | --- |
| Vừa tải xong, chưa thay đổi | Nút tối, không bấm được |
| Có thay đổi hợp lệ, đủ điều kiện thao tác | Nút sáng, cho phép Lưu |
| Sửa rồi trả đúng giá trị đã lưu | Nút tối trở lại |
| Vừa nhấn Lưu | Khóa ngay, hiện “Đang lưu…”, chặn gửi trùng |
| Lưu thành công, không còn thay đổi | Nút tối; hiện “Đã lưu” riêng, không tự bật lại |
| Lưu thất bại | Giữ dữ liệu đang nhập, báo lỗi, cho thử lại nếu dữ liệu còn hợp lệ |

- Tạo helper nhỏ cho snapshot dữ liệu đã lưu, dữ liệu hiện tại, validation và trạng thái gửi; từng form cung cấp cách tạo payload riêng. Không dùng cờ “đã từng sửa” làm tiêu chí duy nhất.
- Ưu tiên mapping MITM, sau đó Cài đặt/Proxy, Plugin, Provider, Combo và cấu hình CLI. Bao phủ cả thao tác thêm/xóa hàng, nút xóa mapping, chọn preset và cập nhật giá trị bằng code.
- Nếu người dùng sửa tiếp trong lúc lưu, chỉ cập nhật snapshot theo payload đã được máy chủ xác nhận; phần sửa mới vẫn là chưa lưu. Xử lý trường mật khẩu rỗng theo nghĩa “giữ bí mật hiện có”, không so sánh với chuỗi đã che.
- Kết hợp dirty với quyền/điều kiện hiện có; không tự mở khóa mapping bị vô hiệu hóa. Các lệnh Tạo mới, Khởi động, Dừng, Reset không áp dụng máy móc quy tắc “phải sửa mới được bấm”.

#### Bước 3 — Làm mượt MITM và tải model

- Dựng khung MITM ổn định; lần đầu chỉ hiện skeleton ở vùng thiếu dữ liệu. Khi refresh, giữ nội dung đang có và báo đang cập nhật tại vùng liên quan, không đổi cả trang thành Loading.
- Tách status, danh sách khóa/model và mappings của công cụ đang chọn. Tái sử dụng dữ liệu tham chiếu trong bộ nhớ phiên, gộp các GET trùng đang chạy; làm mới khi nguồn model/khóa thay đổi hoặc người dùng yêu cầu. Không lưu thêm bí mật vào trình duyệt.
- Đổi công cụ chỉ cập nhật panel mapping; giữ lựa chọn, vị trí cuộn, focus và bản nháp theo công cụ trong phiên trang. Refresh không ghi đè bản nháp; rời trang có thay đổi chưa lưu phải cảnh báo.
- Giữ cơ chế `_toolRun`/`checkPage` đang có và bổ sung bảo vệ theo từng vùng tải để phản hồi cũ không đè công cụ mới. Lỗi tải model chỉ báo tại danh sách model, không làm biến mất toàn bộ trạng thái MITM.
- Sau thao tác, cập nhật đúng phần bị ảnh hưởng; không chạy lại animation xuất hiện trên phần không đổi. Luôn lấy trạng thái mới sau Start/Stop/DNS/CA, không dùng cache cũ để cho phép thao tác nhạy cảm. Giữ cảnh báo cần khởi động lại công cụ sau lưu mapping.
- Chưa đổi backend trong đợt đầu. Nếu vẫn chậm sau tối ưu UI, khảo sát riêng thời gian `/mitm/status`; mã hiện tại có gọi lệnh hệ thống kiểm tra quyền/chứng chỉ, nhưng chưa có số đo để kết luận đây là điểm nghẽn.

### 5.3. Phạm vi và tiêu chí nghiệm thu khi được phép kiểm tra

- Tập trung `app.js`, `tools.js`, `app.css`, `i18n/vi.json` và `i18n/en.json` trong `internal/managementasset/web/`. Giữ giao diện/chức năng hiện tại, không thiết kế lại toàn bộ dashboard hay đổi giao thức proxy.
- MITM: mở lần đầu, đổi công cụ nhanh, refresh, model tải chậm/lỗi; không nháy trắng toàn trang, không mất bản nháp/focus, phản hồi cũ không ghi đè lựa chọn mới.
- Lưu: chưa sửa, sửa rồi hoàn tác, thêm/xóa mapping, bấm liên tiếp, lỗi API, sửa tiếp khi đang lưu; trạng thái nút đúng bảng trên và không gửi trùng.
- Copy: thành công, bấm lại, clipboard bị từ chối và rời trang trong lúc chờ; phản hồi đúng, tồn tại khoảng 2,2 giây, không làm xê dịch bố cục.
- Ưu tiên kiểm tra giao diện với API giả lập hoặc bản đang có khi được người dùng cho phép; không dùng DNS/CA thật để thử hiệu ứng. **Hiện tại không chạy build, test, tạo/chạy EXE mới hoặc tuyên bố các tiêu chí này đã đạt.**
