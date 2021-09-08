(function () {
    'use strict'

    // Fetch all the forms we want to apply custom Bootstrap validation styles to
    let forms = document.querySelectorAll('.needs-validation')

    // Loop over them and prevent submission
    Array.prototype.slice.call(forms)
        .forEach(function (form) {
            form.addEventListener('submit', function (event) {

                if (!form.checkValidity()) {
                    event.preventDefault()
                    event.stopPropagation()
                }

                form.classList.add('was-validated')
            }, false)
        })
})();

function Prompt() {
    let toast = function (c) {
        const {
            msg = '',
            icon = 'success',
            position = 'top-end',
        } = c;

        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
            }
        })

        Toast.fire({});
    }

    let success = function (c) {
        const {
            msg = '',
            title = '',
            footer = '',
        } = c;

        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer,
        })
    }

    let error = function (c) {
        const {
            msg = '',
            title = '',
            footer = '',
        } = c;

        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer,
        })
    }

    let custom = async function custom(c) {
        const {
            icon = '',
            title = '',
            msg = '',
            showConfirmButton = true,
            showCancelButton = true,
            preConfirm = () => {
            },
            willOpen = () => {
            },
            didOpen = () => {
            },
            callback = undefined,
        } = c;

        const {value: result} = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: true,
            focusConfirm: false,
            showCancelButton: showCancelButton,
            showConfirmButton: showConfirmButton,
            preConfirm: preConfirm,
            willOpen: willOpen,
            didOpen: didOpen,
        });

        if (result && callback !== undefined) {
            callback(result);
        }
    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}

function RoomPage(roomID, csrfToken) {
    let attention = Prompt();
    let checkAvailabilityButton = document.getElementById('check-availability-button');

    if (checkAvailabilityButton) {
        checkAvailabilityButton.addEventListener('click', function () {
            const html = `
<div class="container">
    <form id="check-availability" action="" method="POST" novalidate class="needs-validation">
        <div class="row my-3" id="reservation-dates-modal">
            <div class="col-sm-6">
                <label for="startDateModal" class="form-label">Starting date</label>
                <input disabled required type="text" name="start_date" class="form-control" id="startDateModal" placeholder="Arrival">
            </div>
            <div class="col-sm-6">
                <label for="endDateModal" class="form-label">Ending date</label>
                <input disabled required type="text" name="end_date" class="form-control" id="endDateModal" placeholder="Departure">
            </div>
        </div>
    </form>
</div>
`;
            const willOpen = () => {
                    let minDate = new Date();
                    minDate.setDate(minDate.getDate() + 1);

                    const reservationDatesModal = document.getElementById('reservation-dates-modal');
                    const dateRangePicker = new DateRangePicker(reservationDatesModal, {
                        format: 'yyyy-mm-dd',
                        minDate: minDate,
                    });
                },
                didOpen = () => {
                    document.getElementById('startDateModal').removeAttribute('disabled');
                    document.getElementById('endDateModal').removeAttribute('disabled');
                },
                callback = (result) => {
                    console.log(result);

                    let form = document.getElementById('check-availability');
                    let formData = new FormData(form);
                    formData.append('csrf_token', csrfToken);
                    formData.append('room_id', roomID);

                    fetch('/search-availability-json', {
                        method: 'post',
                        body: formData,
                    })
                        .then(response => response.json())
                        .then(data => {
                            if (data.ok) {
                                attention.custom({
                                    icon: 'success',
                                    msg: `<p>Room is available</p><p><a href="/book-room?id=${data.room_id}&s=${data.start_date}&e=${data.end_date}" class="btn btn-primary">Book now!</a></p>`,
                                    showCancelButton: false,
                                    showConfirmButton: false,
                                })
                            } else {
                                attention.error({
                                    msg: 'No availability'
                                })
                            }
                        });
                };

            attention.custom({
                title: 'Choose your dates',
                msg: html,
                willOpen: willOpen,
                didOpen: didOpen,
                callback: callback,
            });
        });
    }
}
