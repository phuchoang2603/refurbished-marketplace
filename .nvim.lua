local function set_file_mark(mark, filepath)
	local abs_path = vim.fn.fnamemodify(filepath, ":p")
	local bufnr = vim.fn.bufadd(abs_path)
	vim.fn.setpos("'" .. mark, { bufnr, 1, 1, 0 })
end

local function clear_global_marks()
	-- Loop through A (65) to Z (90)
	for i = string.byte("A"), string.byte("Z") do
		local mark = string.char(i)
		vim.api.nvim_del_mark(mark)
	end
end
clear_global_marks()

set_file_mark("T", "Tiltfile")

set_file_mark("H", "infra/charts/refurbished-marketplace/values.yaml")
set_file_mark("K", "infra/charts/kafka/values.yaml")
set_file_mark("D", "infra/docker")

set_file_mark("S", "shared/")
set_file_mark("C", "services/cart/internal/service/service.go")
set_file_mark("P", "services/products/internal/service/service.go")
set_file_mark("O", "services/orders/internal/service/service.go")
set_file_mark("U", "services/users/internal/service/service.go")
set_file_mark("I", "services/inventory/internal/service/service.go")
set_file_mark("W", "services/web/internal/handlers/handler.go")
set_file_mark("M", "services/payment/internal/service/service.go")
