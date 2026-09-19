package admin

import (
	"html/template"
	"net/http"

	"presensee/internal/model"
)

var templates = make(map[string]*template.Template)

func init() {
	baseLayout := `
<!DOCTYPE html>
<html lang="id" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Presensee Admin</title>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.12.10/dist/full.min.css" rel="stylesheet" type="text/css" />
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/sweetalert2@11"></script>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap');
        body { font-family: 'Plus Jakarta Sans', sans-serif; }
    </style>
</head>
<body class="bg-base-200 min-h-screen">
    <div class="drawer lg:drawer-open">
        <input id="drawer-toggle" type="checkbox" class="drawer-toggle" />
        <div class="drawer-content flex flex-col">
            <!-- Navbar -->
            <div class="navbar bg-base-100 shadow-sm px-4 lg:px-8 border-b border-base-300">
                <div class="flex-none lg:hidden">
                    <label for="drawer-toggle" class="btn btn-square btn-ghost">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-6 h-6 stroke-current"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
                    </label>
                </div>
                <div class="flex-1">
                    <h1 class="text-xl font-bold text-primary">{{.Title}}</h1>
                </div>
                <div class="flex-none gap-2">
                    <a href="/admin/logout" class="btn btn-sm btn-outline btn-error">Logout</a>
                </div>
            </div>
            <!-- Main Content -->
            <main class="p-4 lg:p-8 flex-1">
                {{template "content" .}}
            </main>
        </div>
        <!-- Sidebar -->
        <div class="drawer-side">
            <label for="drawer-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
            <aside class="bg-base-100 w-64 min-h-full border-r border-base-300 p-4 flex flex-col justify-between">
                <div>
                    <div class="flex items-center gap-3 px-2 py-4 mb-4 border-b border-base-200">
                        <div class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-primary-content font-bold text-xl">P</div>
                        <div>
                            <div class="font-bold text-lg leading-tight">Presensee</div>
                            <div class="text-xs text-base-content/60">Admin Panel (Go)</div>
                        </div>
                    </div>
                    <ul class="menu p-0 gap-1">
                        <li><a href="/admin/dashboard" class="{{if eq .ActiveMenu "dashboard"}}active{{end}} font-medium">📊 Dashboard</a></li>
                        <li><a href="/admin/kelas" class="{{if eq .ActiveMenu "kelas"}}active{{end}} font-medium">🏫 Kelas</a></li>
                        <li><a href="/admin/siswa" class="{{if eq .ActiveMenu "siswa"}}active{{end}} font-medium">👨‍🎓 Siswa</a></li>
                        <li><a href="/admin/users" class="{{if eq .ActiveMenu "users"}}active{{end}} font-medium">👥 Pengguna</a></li>
                        <li><a href="/admin/absensi" class="{{if eq .ActiveMenu "absensi"}}active{{end}} font-medium">📋 Monitoring Absensi</a></li>
                        <li><a href="/admin/kunci" class="{{if eq .ActiveMenu "kunci"}}active{{end}} font-medium">🔒 Kunci Absensi</a></li>
                        <li><a href="/admin/sessions" class="{{if eq .ActiveMenu "sessions"}}active{{end}} font-medium">⏱️ Jadwal Sesi QR</a></li>
                        <div class="divider my-2 text-xs text-base-content/50">TOOLS & DATA</div>
                        <li><a href="/admin/import-siswa" class="{{if eq .ActiveMenu "import"}}active{{end}} font-medium">📥 Import Siswa (.xlsx)</a></li>
                        <li><a href="/admin/export-absensi" class="{{if eq .ActiveMenu "export"}}active{{end}} font-medium">📤 Export Absensi (.xlsx)</a></li>
                        <li><a href="/admin/data-sekolah" class="{{if eq .ActiveMenu "data_sekolah"}}active{{end}} font-medium">⚙️ Data Sekolah</a></li>
                    </ul>
                </div>
                <div class="px-2 py-4 text-xs text-base-content/50 border-t border-base-200">
                    Presensee Go Core v1.0.0
                </div>
            </aside>
        </div>
    </div>
</body>
</html>
`

	// Register all templates
	templates["dashboard"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
    <div class="stat bg-base-100 shadow-sm rounded-2xl border border-base-300">
        <div class="stat-figure text-primary">👨‍🎓</div>
        <div class="stat-title">Total Siswa</div>
        <div class="stat-value text-primary">{{.TotalSiswa}}</div>
    </div>
    <div class="stat bg-base-100 shadow-sm rounded-2xl border border-base-300">
        <div class="stat-figure text-secondary">🏫</div>
        <div class="stat-title">Total Kelas Aktif</div>
        <div class="stat-value text-secondary">{{.TotalKelas}}</div>
    </div>
    <div class="stat bg-base-100 shadow-sm rounded-2xl border border-base-300">
        <div class="stat-figure text-accent">👥</div>
        <div class="stat-title">Total Pengguna</div>
        <div class="stat-value text-accent">{{.TotalUser}}</div>
    </div>
    <div class="stat bg-base-100 shadow-sm rounded-2xl border border-base-300">
        <div class="stat-figure text-success">📋</div>
        <div class="stat-title">Absensi Hari Ini</div>
        <div class="stat-value text-success">{{.TotalAbsensiToday}}</div>
    </div>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 p-6 rounded-2xl">
    <h2 class="text-lg font-bold mb-4">Aksi Cepat</h2>
    <div class="flex flex-wrap gap-3">
        <a href="/admin/import-siswa" class="btn btn-primary">📥 Import Siswa (Excel)</a>
        <a href="/admin/export-absensi" class="btn btn-outline btn-primary">📤 Export Rekap Absensi</a>
        <a href="/admin/kelas/create" class="btn btn-outline">Tambah Kelas</a>
        <a href="/admin/siswa/create" class="btn btn-outline">Tambah Siswa</a>
        <a href="/admin/users/create" class="btn btn-outline">Tambah Pengguna</a>
    </div>
</div>
{{end}}`)

	templates["login"] = template.Must(template.New("login").Parse(`
<!DOCTYPE html>
<html lang="id" data-theme="light">
<head>
    <meta charset="UTF-8">
    <title>Login Admin - Presensee</title>
    <link href="https://cdn.jsdelivr.net/npm/daisyui@4.12.10/dist/full.min.css" rel="stylesheet" type="text/css" />
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-base-200 min-h-screen flex items-center justify-center p-4">
    <div class="card w-full max-w-md bg-base-100 shadow-xl border border-base-300 rounded-3xl p-8">
        <div class="text-center mb-6">
            <div class="inline-flex w-16 h-16 rounded-2xl bg-primary items-center justify-center text-primary-content font-bold text-3xl mb-3">P</div>
            <h1 class="text-2xl font-bold">Presensee Admin</h1>
            <p class="text-sm text-base-content/60">Masuk untuk mengelola sistem absensi</p>
        </div>
        {{if .Error}}
        <div class="alert alert-error mb-4 text-sm py-2">
            <span>{{.Error}}</span>
        </div>
        {{end}}
        <form method="POST" action="/admin/login" class="space-y-4">
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Username</span></label>
                <input type="text" name="username" required autofocus class="input input-bordered w-full" placeholder="admin" />
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Password</span></label>
                <input type="password" name="password" required class="input input-bordered w-full" placeholder="••••••••" />
            </div>
            <button type="submit" class="btn btn-primary w-full mt-2">Masuk Panel Admin</button>
        </form>
    </div>
</body>
</html>
`))

	templates["users_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex justify-between items-center mb-6">
    <div>
        <h2 class="text-2xl font-bold">Daftar Pengguna</h2>
        <p class="text-sm text-base-content/60">Kelola akun wali kelas, sekretaris, kesiswaan, dan guru piket</p>
    </div>
    <a href="/admin/users/create" class="btn btn-primary">➕ Tambah Pengguna</a>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>ID</th>
                    <th>Nama & Username</th>
                    <th>Tipe Akun</th>
                    <th>Role Kelas</th>
                    <th>Hak Akses</th>
                    <th>Aksi</th>
                </tr>
            </thead>
            <tbody>
                {{range .Users}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>
                        <div class="font-bold">{{.DisplayName}}</div>
                        <div class="text-xs text-base-content/50">@{{.Username}}</div>
                    </td>
                    <td>
                        {{if .Type}}
                            <span class="badge badge-primary badge-outline">{{derefUserType .Type}}</span>
                        {{else}}
                            <span class="badge badge-ghost">-</span>
                        {{end}}
                    </td>
                    <td>
                        {{if .WaliKelas}}
                            <span class="badge badge-success badge-sm">Wali: {{.WaliKelas.Name}}</span>
                        {{end}}
                        {{range .SekretarisKelas}}
                            <span class="badge badge-info badge-sm">Sekretaris: {{.Name}}</span>
                        {{end}}
                        {{if and (not .WaliKelas) (eq (len .SekretarisKelas) 0)}}-{{end}}
                    </td>
                    <td>
                        {{if .IsSuperuser}}<span class="badge badge-error badge-sm">Superuser</span>{{end}}
                        {{if .IsStaff}}<span class="badge badge-warning badge-sm">Staff Admin</span>{{end}}
                    </td>
                    <td>
                        <a href="/admin/users/{{.ID}}/edit" class="btn btn-xs btn-outline">Edit</a>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}`)

	templates["user_form"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-2xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-6">{{.Title}}</h2>
    <form method="POST" class="space-y-4">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Username</span></label>
            <input type="text" name="username" value="{{.User.Username}}" required class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Nama Lengkap</span></label>
            <input type="text" name="full_name" value="{{.User.FullName}}" class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Password {{if not .IsNew}}(Kosongkan jika tidak ingin mengubah){{end}}</span></label>
            <input type="password" name="password" {{if .IsNew}}required{{end}} class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Tipe Pengguna</span></label>
            <select name="type" id="user-type" class="select select-bordered">
                <option value="----">----</option>
                <option value="wali_kelas" {{if and .User.Type (eq (derefUserType .User.Type) "wali_kelas")}}selected{{end}}>Wali Kelas</option>
                <option value="sekretaris" {{if and .User.Type (eq (derefUserType .User.Type) "sekretaris")}}selected{{end}}>Sekretaris</option>
                <option value="kesiswaan" {{if and .User.Type (eq (derefUserType .User.Type) "kesiswaan")}}selected{{end}}>Kesiswaan</option>
                <option value="guru_piket" {{if and .User.Type (eq (derefUserType .User.Type) "guru_piket")}}selected{{end}}>Guru Piket</option>
            </select>
        </div>
        <div class="form-control" id="kelas-group">
            <label class="label"><span class="label-text font-medium">Kelas Terkait</span></label>
            <select name="kelas_id" class="select select-bordered">
                <option value="----">----</option>
                {{range .Kelass}}
                <option value="{{.ID}}" {{if eq $.SelectedKelasID .ID}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="divider">Hak Akses</div>
        <div class="flex gap-6">
            <label class="cursor-pointer label gap-2">
                <input type="checkbox" name="is_staff" {{if .User.IsStaff}}checked{{end}} class="checkbox checkbox-primary" />
                <span class="label-text">Akses Admin Panel</span>
            </label>
            <label class="cursor-pointer label gap-2">
                <input type="checkbox" name="is_superuser" {{if .User.IsSuperuser}}checked{{end}} class="checkbox checkbox-primary" />
                <span class="label-text">Superuser (Admin Utama)</span>
            </label>
        </div>
        <div class="flex justify-end gap-3 mt-8">
            <a href="/admin/users" class="btn btn-ghost">Batal</a>
            <button type="submit" class="btn btn-primary">Simpan Pengguna</button>
        </div>
    </form>
</div>
{{end}}`)

	templates["kelas_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex justify-between items-center mb-6">
    <div>
        <h2 class="text-2xl font-bold">Daftar Kelas</h2>
        <p class="text-sm text-base-content/60">Kelola kelas, wali kelas, dan kenaikan kelas</p>
    </div>
    <a href="/admin/kelas/create" class="btn btn-primary">➕ Tambah Kelas</a>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>ID</th>
                    <th>Nama Kelas</th>
                    <th>Wali Kelas</th>
                    <th>Sekretaris</th>
                    <th>Status</th>
                    <th>Aksi</th>
                </tr>
            </thead>
            <tbody>
                {{range .Kelass}}
                <tr>
                    <td>{{.ID}}</td>
                    <td class="font-bold">{{.Name}}</td>
                    <td>
                        {{if .WaliKelas}}{{.WaliKelas.DisplayName}}{{else}}<span class="text-base-content/40">-</span>{{end}}
                    </td>
                    <td>
                        {{range .Sekretaris}}
                            <span class="badge badge-sm badge-outline">{{.DisplayName}}</span>
                        {{else}}
                            <span class="text-base-content/40">-</span>
                        {{end}}
                    </td>
                    <td>
                        {{if .Active}}
                            <span class="badge badge-success badge-sm">Aktif</span>
                        {{else}}
                            <span class="badge badge-ghost badge-sm">Tidak Aktif</span>
                        {{end}}
                    </td>
                    <td class="flex gap-2">
                        <a href="/admin/kelas/{{.ID}}/edit" class="btn btn-xs btn-outline">Edit</a>
                        <button onclick="triggerNaikKelas({{.ID}}, '{{.Name}}')" class="btn btn-xs btn-primary btn-outline">Naik Kelas</button>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
<script>
function triggerNaikKelas(oldKelasId, oldKelasName) {
    Swal.fire({
        title: 'Naik Kelas: ' + oldKelasName,
        text: 'Masukkan nama kelas baru untuk memindahkan seluruh siswa:',
        input: 'text',
        showCancelButton: true,
        confirmButtonText: 'Proses Naik Kelas',
        cancelButtonText: 'Batal',
        showLoaderOnConfirm: true,
        preConfirm: async (newKelasName) => {
            if (!newKelasName) {
                Swal.showValidationMessage('Nama kelas baru wajib diisi');
                return;
            }
            try {
                const res = await fetch('/admin/naik-kelas', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ old_kelas_id: oldKelasId, new_kelas_name: newKelasName })
                });
                const txt = await res.text();
                if (txt === 'NOT_OK') {
                    throw new Error('Gagal memproses naik kelas');
                }
                return txt;
            } catch (err) {
                Swal.showValidationMessage(err.message);
            }
        }
    }).then((result) => {
        if (result.isConfirmed) {
            Swal.fire('Berhasil!', 'Siswa berhasil dipindahkan ke kelas baru.', 'success').then(() => {
                window.location.reload();
            });
        }
    });
}
</script>
{{end}}`)

	templates["kelas_form"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-6">{{.Title}}</h2>
    <form method="POST" class="space-y-4">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Nama Kelas</span></label>
            <input type="text" name="name" value="{{.Kelas.Name}}" required class="input input-bordered" placeholder="Contoh: XII IPA 1" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Wali Kelas</span></label>
            <select name="wali_kelas_id" class="select select-bordered">
                <option value="0">---- Tidak Ada ----</option>
                {{range .WaliUsers}}
                <option value="{{.ID}}" {{if and $.Kelas.WaliKelasID (eq (derefUint $.Kelas.WaliKelasID) .ID)}}selected{{end}}>{{.DisplayName}} (@{{.Username}})</option>
                {{end}}
            </select>
        </div>
        <div class="form-control">
            <label class="cursor-pointer label gap-2 justify-start">
                <input type="checkbox" name="active" {{if .Kelas.Active}}checked{{end}} class="checkbox checkbox-primary" />
                <span class="label-text font-medium">Kelas Aktif</span>
            </label>
        </div>
        <div class="flex justify-end gap-3 mt-8">
            <a href="/admin/kelas" class="btn btn-ghost">Batal</a>
            <button type="submit" class="btn btn-primary">Simpan Kelas</button>
        </div>
    </form>
</div>
{{end}}`)

	templates["siswa_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6">
    <div>
        <h2 class="text-2xl font-bold">Daftar Siswa</h2>
        <p class="text-sm text-base-content/60">Kelola data induk siswa</p>
    </div>
    <div class="flex gap-2">
        <a href="/admin/kartu-siswa{{if .SelectedKelasID}}?kelas_id={{.SelectedKelasID}}{{end}}" target="_blank" class="btn btn-outline btn-secondary">🪪 Cetak Kartu Siswa</a>
        <a href="/admin/siswa/create" class="btn btn-primary">➕ Tambah Siswa</a>
    </div>
</div>
<!-- Filter Card -->
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-4 mb-6">
    <form method="GET" class="flex flex-wrap gap-4 items-end">
        <div class="form-control flex-1 min-w-[200px]">
            <label class="label py-1"><span class="label-text text-xs">Filter Kelas</span></label>
            <select name="kelas_id" class="select select-bordered select-sm w-full">
                <option value="">Semua Kelas</option>
                {{range .Kelass}}
                <option value="{{.ID}}" {{if eq $.SelectedKelasID (printf "%d" .ID)}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="form-control flex-1 min-w-[200px]">
            <label class="label py-1"><span class="label-text text-xs">Cari Nama / NIS</span></label>
            <input type="text" name="search" value="{{.Search}}" placeholder="Ketik nama atau NIS..." class="input input-bordered input-sm w-full" />
        </div>
        <button type="submit" class="btn btn-sm btn-primary">Filter</button>
        <a href="/admin/siswa" class="btn btn-sm btn-ghost">Reset</a>
    </form>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>ID</th>
                    <th>Nama Lengkap</th>
                    <th>Kelas</th>
                    <th>NIS</th>
                    <th>NISN</th>
                    <th>Aksi</th>
                </tr>
            </thead>
            <tbody>
                {{range .Siswas}}
                <tr>
                    <td>{{.ID}}</td>
                    <td class="font-bold">{{.FullName}}</td>
                    <td><span class="badge badge-outline">{{.Kelas.Name}}</span></td>
                    <td>{{if .NIS}}{{.NIS}}{{else}}-{{end}}</td>
                    <td>{{if .NISN}}{{.NISN}}{{else}}-{{end}}</td>
                    <td>
                        <a href="/admin/siswa/{{.ID}}/edit" class="btn btn-xs btn-outline">Edit</a>
                    </td>
                </tr>
                {{else}}
                <tr><td colspan="6" class="text-center py-8 text-base-content/50">Tidak ada data siswa ditemukan</td></tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}`)

	templates["siswa_form"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-6">{{.Title}}</h2>
    <form method="POST" class="space-y-4">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Nama Lengkap Siswa</span></label>
            <input type="text" name="full_name" value="{{.Siswa.FullName}}" required class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Kelas</span></label>
            <select name="kelas_id" required class="select select-bordered">
                {{range .Kelass}}
                <option value="{{.ID}}" {{if eq $.Siswa.KelasID .ID}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">NIS (Nomor Induk Siswa)</span></label>
            <input type="text" name="nis" value="{{.Siswa.NIS}}" class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">NISN</span></label>
            <input type="text" name="nisn" value="{{.Siswa.NISN}}" class="input input-bordered" />
        </div>
        <div class="flex justify-end gap-3 mt-8">
            <a href="/admin/siswa" class="btn btn-ghost">Batal</a>
            <button type="submit" class="btn btn-primary">Simpan Siswa</button>
        </div>
    </form>
</div>
{{end}}`)

	templates["absensi_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex justify-between items-center mb-6">
    <div>
        <h2 class="text-2xl font-bold">Monitoring Absensi</h2>
        <p class="text-sm text-base-content/60">Lihat rekaman absensi harian per kelas</p>
    </div>
</div>
<!-- Filter Card -->
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-4 mb-6">
    <form method="GET" class="flex flex-wrap gap-4 items-end">
        <div class="form-control">
            <label class="label py-1"><span class="label-text text-xs">Pilih Tanggal</span></label>
            <input type="date" name="date" value="{{.SelectedDate}}" class="input input-bordered input-sm" />
        </div>
        <div class="form-control">
            <label class="label py-1"><span class="label-text text-xs">Pilih Kelas</span></label>
            <select name="kelas_id" class="select select-bordered select-sm">
                <option value="">Semua Kelas</option>
                {{range .Kelass}}
                <option value="{{.ID}}" {{if eq $.SelectedKelasID (printf "%d" .ID)}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <button type="submit" class="btn btn-sm btn-primary">Tampilkan</button>
    </form>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>Siswa</th>
                    <th>Kelas</th>
                    <th>Status Absensi</th>
                    <th>Diabsen Oleh</th>
                    <th>Waktu Update</th>
                </tr>
            </thead>
            <tbody>
                {{range .Absensies}}
                <tr>
                    <td class="font-bold">{{.Siswa.FullName}}</td>
                    <td>{{.Siswa.Kelas.Name}}</td>
                    <td>
                        {{if eq .Status "hadir"}}<span class="badge badge-success badge-sm">Hadir</span>{{end}}
                        {{if eq .Status "sakit"}}<span class="badge badge-warning badge-sm">Sakit</span>{{end}}
                        {{if eq .Status "izin"}}<span class="badge badge-info badge-sm">Izin</span>{{end}}
                        {{if eq .Status "alfa"}}<span class="badge badge-error badge-sm">Alfa</span>{{end}}
                        {{if eq .Status "bolos"}}<span class="badge badge-secondary badge-sm">Bolos</span>{{end}}
                        {{if eq .Status "tunggu"}}<span class="badge badge-ghost badge-sm">Menunggu</span>{{end}}
                    </td>
                    <td>{{if .By}}{{.By.DisplayName}}{{else}}-{{end}}</td>
                    <td class="text-xs text-base-content/60">{{.UpdatedAt.Format "15:04:05"}}</td>
                </tr>
                {{else}}
                <tr><td colspan="5" class="text-center py-8 text-base-content/50">Belum ada data absensi untuk tanggal ini</td></tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}`)

	templates["kunci_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex justify-between items-center mb-6">
    <div>
        <h2 class="text-2xl font-bold">Kunci Absensi</h2>
        <p class="text-sm text-base-content/60">Status kunci absensi per kelas dan tanggal</p>
    </div>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>Kelas</th>
                    <th>Tanggal</th>
                    <th>Status</th>
                    <th>Aksi</th>
                </tr>
            </thead>
            <tbody>
                {{range .Kuncis}}
                <tr>
                    <td class="font-bold">{{.Kelas.Name}}</td>
                    <td>{{.Date.Format "02 Jan 2006"}}</td>
                    <td>
                        {{if .Locked}}
                            <span class="badge badge-error badge-sm">🔒 Terkunci</span>
                        {{else}}
                            <span class="badge badge-success badge-sm">🔓 Terbuka</span>
                        {{end}}
                    </td>
                    <td>
                        <form method="POST" action="/admin/kunci/toggle" class="inline">
                            <input type="hidden" name="kelas_id" value="{{.KelasID}}" />
                            <input type="hidden" name="date" value="{{.Date.Format "2006-01-02"}}" />
                            <input type="hidden" name="locked" value="{{if .Locked}}false{{else}}true{{end}}" />
                            <button type="submit" class="btn btn-xs {{if .Locked}}btn-success{{else}}btn-error{{end}} btn-outline">
                                {{if .Locked}}Buka Kunci{{else}}Kunci Sekarang{{end}}
                            </button>
                        </form>
                    </td>
                </tr>
                {{else}}
                <tr><td colspan="4" class="text-center py-8 text-base-content/50">Belum ada catatan kunci absensi</td></tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}`)

	templates["sessions_list"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="flex justify-between items-center mb-6">
    <div>
        <h2 class="text-2xl font-bold">Jadwal Sesi Absensi (QR)</h2>
        <p class="text-sm text-base-content/60">Atur jam masuk, jam pulang, dan hari untuk pemindaian QR</p>
    </div>
    <a href="/admin/sessions/create" class="btn btn-primary">➕ Tambah Sesi</a>
</div>
<div class="card bg-base-100 shadow-sm border border-base-300 rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
        <table class="table table-zebra">
            <thead class="bg-base-200/50">
                <tr>
                    <th>Hari Aktif</th>
                    <th>Jam Masuk</th>
                    <th>Jam Pulang</th>
                    <th>Kelas Terkait</th>
                    <th>Aksi</th>
                </tr>
            </thead>
            <tbody>
                {{range .Sessions}}
                <tr>
                    <td>
                        <div class="flex flex-wrap gap-1">
                            {{if .Senin}}<span class="badge badge-sm badge-outline">Senin</span>{{end}}
                            {{if .Selasa}}<span class="badge badge-sm badge-outline">Selasa</span>{{end}}
                            {{if .Rabu}}<span class="badge badge-sm badge-outline">Rabu</span>{{end}}
                            {{if .Kamis}}<span class="badge badge-sm badge-outline">Kamis</span>{{end}}
                            {{if .Jumat}}<span class="badge badge-sm badge-outline">Jumat</span>{{end}}
                            {{if .Sabtu}}<span class="badge badge-sm badge-outline">Sabtu</span>{{end}}
                        </div>
                    </td>
                    <td class="font-mono font-bold">{{.JamMasuk}}</td>
                    <td class="font-mono font-bold">{{.JamKeluar}}</td>
                    <td>
                        <div class="flex flex-wrap gap-1">
                            {{range .Kelas}}<span class="badge badge-sm badge-primary">{{.Name}}</span>{{end}}
                        </div>
                    </td>
                    <td>
                        <a href="/admin/sessions/{{.ID}}/edit" class="btn btn-xs btn-outline">Edit</a>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}`)

	templates["session_form"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-2xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-6">{{.Title}}</h2>
    <form method="POST" class="space-y-6">
        <div>
            <label class="label"><span class="label-text font-medium">Hari Aktif</span></label>
            <div class="grid grid-cols-3 md:grid-cols-6 gap-2">
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Senin</span>
                    <input type="checkbox" name="senin" {{if .Session.Senin}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Selasa</span>
                    <input type="checkbox" name="selasa" {{if .Session.Selasa}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Rabu</span>
                    <input type="checkbox" name="rabu" {{if .Session.Rabu}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Kamis</span>
                    <input type="checkbox" name="kamis" {{if .Session.Kamis}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Jumat</span>
                    <input type="checkbox" name="jumat" {{if .Session.Jumat}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
                <label class="cursor-pointer label border border-base-300 rounded-xl p-3 flex flex-col items-center gap-2">
                    <span class="label-text font-bold">Sabtu</span>
                    <input type="checkbox" name="sabtu" {{if .Session.Sabtu}}checked{{end}} class="checkbox checkbox-primary" />
                </label>
            </div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Jam Masuk (Mulai)</span></label>
                <input type="text" name="jam_masuk" value="{{.Session.JamMasuk}}" placeholder="07:00" required class="input input-bordered font-mono" />
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Jam Pulang (Mulai)</span></label>
                <input type="text" name="jam_keluar_mulai_absen" value="{{.Session.JamKeluarMulaiAbsen}}" placeholder="13:30" class="input input-bordered font-mono" />
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Jam Keluar (Batas Akhir)</span></label>
                <input type="text" name="jam_keluar" value="{{.Session.JamKeluar}}" placeholder="14:00" required class="input input-bordered font-mono" />
            </div>
        </div>
        <div>
            <label class="label"><span class="label-text font-medium">Pilih Kelas yang Menerapkan Jadwal Ini</span></label>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-2 border border-base-300 p-4 rounded-xl max-h-48 overflow-y-auto">
                {{range .Kelass}}
                <label class="cursor-pointer label justify-start gap-2">
                    <input type="checkbox" name="kelas_ids" value="{{.ID}}" class="checkbox checkbox-primary checkbox-sm" />
                    <span class="label-text">{{.Name}}</span>
                </label>
                {{end}}
            </div>
        </div>
        <div class="flex justify-end gap-3 mt-8">
            <a href="/admin/sessions" class="btn btn-ghost">Batal</a>
            <button type="submit" class="btn btn-primary">Simpan Sesi Jadwal</button>
        </div>
    </form>
</div>
{{end}}`)

	templates["data_sekolah"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-6">Data & Pengaturan Sekolah</h2>
    <form method="POST" class="space-y-4">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Nama Aplikasi</span></label>
            <input type="text" name="nama_aplikasi" value="{{.Data.NamaAplikasi}}" required class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Nama Sekolah</span></label>
            <input type="text" name="nama_sekolah" value="{{.Data.NamaSekolah}}" required class="input input-bordered" />
        </div>
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Deskripsi Sekolah</span></label>
            <textarea name="deskripsi_sekolah" class="textarea textarea-bordered h-24">{{.Data.DeskripsiSekolah}}</textarea>
        </div>
        <div class="flex justify-end gap-3 mt-8">
            <button type="submit" class="btn btn-primary">Simpan Pengaturan</button>
        </div>
    </form>
</div>
{{end}}`)

	templates["import_siswa"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-2">Import Siswa dari Excel</h2>
    <p class="text-sm text-base-content/60 mb-6">Unggah berkas spreadsheet (.xlsx) dengan kolom: Nama, Kelas, NIS, NISN</p>
    {{if .SuccessMsg}}
    <div class="alert alert-success mb-6">
        <span>{{.SuccessMsg}}</span>
    </div>
    {{end}}
    <form method="POST" enctype="multipart/form-data" class="space-y-6">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Pilih File .xlsx</span></label>
            <input type="file" name="excel_file" accept=".xlsx" required class="file-input file-input-bordered w-full" />
        </div>
        <button type="submit" class="btn btn-primary w-full">Unggah & Proses Import</button>
    </form>
</div>
{{end}}`)

	templates["export_absensi"] = parseTmpl(baseLayout, `{{define "content"}}
<div class="max-w-xl mx-auto card bg-base-100 shadow-sm border border-base-300 rounded-2xl p-6 lg:p-8">
    <h2 class="text-2xl font-bold mb-2">Export Rekap Absensi</h2>
    <p class="text-sm text-base-content/60 mb-6">Unduh matriks rekap absensi bulanan dalam format Excel (.xlsx)</p>
    <form method="POST" class="space-y-4">
        <div class="form-control">
            <label class="label"><span class="label-text font-medium">Pilih Kelas</span></label>
            <select name="kelas_id" required class="select select-bordered w-full">
                {{range .Kelass}}
                <option value="{{.ID}}">{{.Name}}</option>
                {{end}}
            </select>
        </div>
        <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Bulan</span></label>
                <select name="month" class="select select-bordered w-full">
                    <option value="1">Januari</option>
                    <option value="2">Februari</option>
                    <option value="3">Maret</option>
                    <option value="4">April</option>
                    <option value="5">Mei</option>
                    <option value="6">Juni</option>
                    <option value="7">Juli</option>
                    <option value="8">Agustus</option>
                    <option value="9">September</option>
                    <option value="10">Oktober</option>
                    <option value="11">November</option>
                    <option value="12">Desember</option>
                </select>
            </div>
            <div class="form-control">
                <label class="label"><span class="label-text font-medium">Tahun</span></label>
                <input type="number" name="year" value="{{.CurrentYear}}" class="input input-bordered w-full" />
            </div>
        </div>
        <button type="submit" class="btn btn-primary w-full mt-4">📥 Unduh Excel (.xlsx)</button>
    </form>
</div>
{{end}}`)

	templates["kartu_siswa"] = template.Must(template.New("kartu_siswa").Parse(`
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Kartu Siswa</title>
    <style>
        @page { size: A4; margin: 10mm; }
        body { font-family: sans-serif; margin: 0; background: #f0f0f0; }
        .grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 15px; }
        .card { width: 85.6mm; height: 53.98mm; background: white; border: 1px solid #ccc; border-radius: 8px; padding: 12px; box-sizing: border-box; display: flex; flex-direction: column; justify-content: space-between; page-break-inside: avoid; }
        .header { font-size: 11px; font-weight: bold; border-bottom: 2px solid #2563eb; padding-bottom: 4px; }
        .name { font-size: 14px; font-weight: bold; margin-top: 6px; }
        .info { font-size: 10px; color: #555; }
        @media print { body { background: none; } }
    </style>
</head>
<body>
    <div class="grid">
        {{range .Siswas}}
        <div class="card">
            <div class="header">{{$.DataSekolah.NamaSekolah}}</div>
            <div>
                <div class="name">{{.FullName}}</div>
                <div class="info">Kelas: {{.Kelas.Name}}</div>
                <div class="info">NIS: {{.NIS}} | NISN: {{.NISN}}</div>
            </div>
            <div class="info text-right">KARTU PELAJAR</div>
        </div>
        {{end}}
    </div>
</body>
</html>
`))
}

func parseTmpl(base, content string) *template.Template {
	t := template.New("base").Funcs(template.FuncMap{
		"containsUint": func(slice []uint, val uint) bool {
			for _, item := range slice {
				if item == val {
					return true
				}
			}
			return false
		},
		"derefUserType": func(t any) string {
			if t == nil {
				return "-"
			}
			return fmtSprint(t)
		},
		"derefUint": func(u *uint) uint {
			if u == nil {
				return 0
			}
			return *u
		},
	})
	template.Must(t.Parse(base))
	template.Must(t.Parse(content))
	return t
}

func renderTemplate(w http.ResponseWriter, name string, data any) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "Template not found: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func fmtSprint(a any) string {
	if str, ok := a.(*model.UserType); ok && str != nil {
		return string(*str)
	}
	if str, ok := a.(model.UserType); ok {
		return string(str)
	}
	return ""
}
